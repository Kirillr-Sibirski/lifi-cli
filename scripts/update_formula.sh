#!/usr/bin/env bash

set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "usage: $0 vX.Y.Z" >&2
  exit 1
fi

version="${1#v}"
tag="v${version}"
repo="https://github.com/Kirillr-Sibirski/lifi-cli"
url="${repo}/archive/refs/tags/${tag}.tar.gz"
formula="Formula/lifi.rb"

if [[ ! -f "${formula}" ]]; then
  echo "missing ${formula}" >&2
  exit 1
fi

tmp="$(mktemp -t lifi-formula.XXXXXX.tar.gz)"
trap 'rm -f "${tmp}"' EXIT

curl -sS -L -o "${tmp}" "${url}"
sha="$(shasum -a 256 "${tmp}" | awk '{print $1}')"

perl -0pi -e 's{url "https://github.com/Kirillr-Sibirski/lifi-cli/archive/refs/tags/v[^"]+\.tar\.gz"}{url "https://github.com/Kirillr-Sibirski/lifi-cli/archive/refs/tags/'"${tag}"'\.tar\.gz"}g' "${formula}"
perl -0pi -e 's{sha256 "[^"]+"}{sha256 "'"${sha}"'"}g' "${formula}"

echo "Updated ${formula} to ${tag}"
echo "sha256 ${sha}"
