#!/bin/bash

set -euo pipefail
set -x

go run . -channel "stable" -out "./static/stable.json"
go run . -channel "regular" -out "./static/regular.json"
go run . -channel "rapid" -out "./static/rapid.json"
