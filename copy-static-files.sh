#!/bin/bash
set -e
SCRIPT_DIR=$(dirname $(readlink -f $0))
if [[ -z ${GOPATH} ]]; then
  echo "GOPATH not defined" >&2
fi
if [[ ! -d "${SCRIPT_DIR}/server/static" ]]; then
  mkdir -p "${SCRIPT_DIR}/server/static"
fi
cp "${SCRIPT_DIR}/server/static/"* "${GOPATH}/bin/static/"
