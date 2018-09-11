#!/bin/bash
set -e
if [[ -z ${GOPATH} ]]; then
  echo "GOPATH not defined" >&2
fi
"${GOPATH}/bin/contentscraper.exe" -enable-harvest=false -port=8191 -log-level=DEBUG
