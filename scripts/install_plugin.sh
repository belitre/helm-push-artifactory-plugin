#!/bin/sh -e

# Copied w/ love from the excellent hypnoglow/helm-s3

if [ -n "${HELM_PUSH_PLUGIN_NO_INSTALL_HOOK}" ]; then
    echo "Development mode: not downloading versioned release."
    exit 0
fi

version="$(cat plugin.yaml | grep "version" | cut -d '"' -f 2)"
echo "Downloading and installing helm-push-artifactory v${version} ..."

url=""
osname=""

if [ "$(uname)" = "Darwin" ]; then
    osname="darwin-amd64"
elif [ "$(uname)" = "Linux" ] ; then
    osname="linux-amd64"
else
    echo "Windows not supported..."
    exit 0
fi

url="https://github.com/belitre/helm-push-artifactory-plugin/releases/download/v${version}/helm-push-artifactory-v${version}-${osname}.tar.gz"

echo $url

mkdir -p "bin"
mkdir -p "releases/v${version}"

# Download with curl if possible.
if [ -x "$(which curl 2>/dev/null)" ]; then
    curl -sSL "${url}" -o "releases/v${version}.tar.gz"
else
    wget -q "${url}" -O "releases/v${version}.tar.gz"
fi
tar xzf "releases/v${version}.tar.gz" -C "releases/v${version}"
mv "releases/v${version}/${osname}/helm-push-artifactory" "bin/helm-push-artifactory"

