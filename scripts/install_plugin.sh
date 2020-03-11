#!/bin/sh -e

# Copied w/ love from the excellent hypnoglow/helm-s3

if [ -n "${HELM_PUSH_PLUGIN_NO_INSTALL_HOOK}" ]; then
    echo "Development mode: not downloading versioned release."
    exit 0
fi

supported_helm_version="v2"

helm_version="$(helm version | grep "Version" | cut -d '"' -f 2 | cut -d '.' -f 1)"

if [ "$helm_version" != "$supported_helm_version" ]; then
    echo "Error, this version of helm-push-artifactory-plugin is only for Helm v3"
    echo "Please use --version v1.0.1 or higher to install the plugin"
    exit 1
fi

echo "Installing plugin for Helm $supported_helm_version..."

version="$(cat plugin.yaml | grep "version" | cut -d '"' -f 2)"
echo "Downloading and installing helm-push-artifactory v${version} ..."

url=""
osname=""

if [ "$(uname)" = "Darwin" ]; then
    osname="darwin-amd64"
elif [ "$(uname)" = "Linux" ] ; then
    osname="linux-amd64"
else
    echo "Windows not supported... please check the README"
    exit 0
fi

filename="helm-push-artifactory-v${version}-${osname}.tar.gz"

url="https://github.com/belitre/helm-push-artifactory-plugin/releases/download/v${version}/${filename}"

echo $url

mkdir -p "releases/v${version}"

if [ -n "${HELM_PUSH_PLUGIN_LOCAL_VERSION}" ]; then
    cp -f ./_dist/${filename} releases/v${version}.tar.gz
else
# Download with curl if possible.
if [ -x "$(which curl 2>/dev/null)" ]; then
    curl -sSL "${url}" -o "releases/v${version}.tar.gz"
else
    wget -q "${url}" -O "releases/v${version}.tar.gz"
fi
fi

tar xzf "releases/v${version}.tar.gz" -C "releases/v${version}"
cp -rf releases/v${version}/helm-push-artifactory-plugin/* ./
