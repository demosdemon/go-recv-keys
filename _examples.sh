#!/usr/bin/env bash

set -eux

go build

node() {
	declare -a KEYSERVERS=(
		hkp://p80.pool.sks-keyservers.net:80
		hkp://ipv4.pool.sks-keyservers.net
		hkp://pgp.mit.edu:80
	)

	declare -a NODE_KEYS=(
		94AE36675C464D64BAFA68DD7434390BDBE9B9C5
		FD3A5288F042B6850C66B31F09FE44734EB7990E
		71DCFD284A79C3B38668286BC97EC7A07EDE3FC1
		DD8F2338BAE7501E3DD5AC78C273792F7D83545D
		C4F0DFFF4E8C1A8236409D08E73BC641CC11F4C8
		B9AE9905FFD7803F25714661B63B535A4C206CA9
		77984A986EBC2AA786BC0F66B01FBB92821C587A
		8FCCA13FEF1D0C2E91008E09770F7A9A5AE15600
		4ED778F539E3634C779C87C6D7062848A1AB005C
		A48C2BEE680E841632CD4E44F07496B3EB3C1762
		B9E2F5981AA6E0CD28160D9FF13993A75599653C
	)

	declare -a args=()
	for ks in "${KEYSERVERS[@]}"; do
		args+=(--keyserver "$ks")
	done

	args+=("${NODE_KEYS[@]}")

	./go-recv-keys --yaml "${args[@]}" 5083B8A58814CE1194AB82B22A8348B169ABD965 8328D8B25A64B8C5BBC3BA23FD50D44C03D3174C
}

node
