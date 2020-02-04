package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/spf13/pflag"
	"golang.org/x/crypto/openpgp"
	"gopkg.in/yaml.v2"
)

var httpClient = http.DefaultClient

func main() {
	log.SetOutput(os.Stderr)
	log.SetPrefix("go-recv-keys: ")

	var (
		flagTimeout    time.Duration
		flagKeyServers []string
		flagJSON       bool
		flagYAML       bool
		flagImport     bool
	)
	pflag.DurationVar(&flagTimeout, "timeout", 30*time.Second, "The key server timeout duration.")
	pflag.StringSliceVarP(&flagKeyServers, "keyserver", "k", nil, "The key server to query (multiple allowed).")
	pflag.BoolVar(&flagJSON, "json", false, "Output as JSON")
	pflag.BoolVar(&flagYAML, "yaml", false, "Output as YAML")
	pflag.BoolVar(&flagImport, "import", false, "Execute `gpg --batch --import` with results.")
	pflag.Parse()

	if flagJSON && flagYAML {
		log.Fatal("--json and --yaml are mutually exclusive")
	}

	outputText := !flagJSON && !flagYAML && !flagImport

	keyservers := make([]*KeyServer, 0, len(flagKeyServers))
	for _, ks := range flagKeyServers {
		ks, err := NewKeyServer(ks)
		if err != nil {
			log.Printf("invalid keyserver: %v", err)
			continue
		}
		keyservers = append(keyservers, ks)
	}

	if len(keyservers) == 0 {
		log.Fatal("no valid keyservers provided")
	}

	args := pflag.Args()
	keys := make([]*Key, 0, len(args))
	for _, k := range args {
		k, err := NewKey(k)
		if err != nil {
			log.Printf("invalid key: %v", err)
			continue
		}
		keys = append(keys, k)
	}

	if len(keys) == 0 {
		log.Fatal("no valid keys provided")
	}

	res := magic(context.Background(), flagTimeout, keys, keyservers)

	if flagJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(res); err != nil {
			log.Fatalf("error encoding json: %v", err)
		}
		_, _ = os.Stdout.WriteString("\n")
	} else if flagYAML {
		enc := yaml.NewEncoder(os.Stdout)
		if err := enc.Encode(res); err != nil {
			log.Fatalf("error encoding yaml: %v", err)
		}
		_, _ = os.Stdout.WriteString("\n")
	}

	if !(outputText || flagImport) {
		return
	}

	el := make(openpgp.EntityList, 0, len(res))
	for _, r := range res {
		if r.Error != nil {
			log.Printf("error fetching key: %#v", r)
			continue
		}
		el = append(el, r.Result...)
	}

	if outputText {
		if err := serializeEntityList(el, os.Stdout); err != nil {
			log.Fatalf("error serializing output: %v", err)
		}
	}

	if flagImport {
		cmd := exec.Command("gpg", "--batch", "--import")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		stdin, err := cmd.StdinPipe()
		if err != nil {
			log.Fatalf("error allocating input pipe for subcommand: %v", err)
		}

		if err := cmd.Start(); err != nil {
			_ = stdin.Close()
			log.Fatalf("error starting subcommand: %v", err)
		}

		if err := serializeEntityList(el, stdin); err != nil {
			_ = stdin.Close()
			log.Fatalf("error serializing output to subcommand: %v", err)
		}

		if err := stdin.Close(); err != nil {
			log.Fatalf("error closing stdin pipe to subcommand: %v", err)
		}

		if err := cmd.Wait(); err != nil {
			log.Fatalf("error waiting for subcommand to finish: %v", err)
		}
	}
}

func magic(ctx context.Context, timeout time.Duration, keys []*Key, keyservers []*KeyServer) []*Result {
	l := len(keys)

	var wg sync.WaitGroup
	wg.Add(l)

	rv := make([]*Result, l)
	for idx, k := range keys {
		go func(idx int, k *Key) {
			defer wg.Done()
			rv[idx] = k.Resolve(ctx, timeout, keyservers)
		}(idx, k)
	}

	wg.Wait()
	return rv
}
