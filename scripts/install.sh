#!/usr/bin/env bash
set -e

# Determine the platform using 'OS' and 'ARCH'
initPlatform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    if [ $OS == "darwin" ] && [ $ARCH == "arm64" ]; then
        PLATFORM="darwin_arm64"
    elif [ $OS == "darwin" ] && [ $ARCH == "x86_64" ]; then
        PLATFORM="darwin_amd64"
    elif [ $OS == "linux" ] && [ $ARCH == "x86_64" ]; then
        PLATFORM="linux_amd64"
    else
        printf "Platform not supported: os=$OS arch=$ARCH\n"
        exit 1
    fi
}

# Create a temporary directory and setup deletion on script exit using the 'EXIT' signal
initTmpDir() {
    TMP_DIR=$(mktemp -d)
    trap 'rm -rf -- "${TMP_DIR}"' EXIT
    cd $TMP_DIR
}

# Ensure that dependency is installed and executable, exit and print help message if not
checkDependency() {
    if ! [ -x "$(command -v $1)" ]; then
        printf "'$1' could not be found, this script depends on it, please install and try again.\n"
        exit 1
    fi
}

# Download the binary and check the integrity using the SHA256 checksum
downloadBinary() {
    CDN="cdn.rilldata.com"
    LATEST_URL="https://${CDN}/rill/latest.txt"
    if [ "${VERSION}" == "latest" ]; then
        VERSION=$(curl --silent ${LATEST_URL})
    fi
    BINARY_URL="https://${CDN}/rill/${VERSION}/rill_${PLATFORM}.zip"
    CHECKSUM_URL="https://${CDN}/rill/${VERSION}/checksums.txt"

    printf "Downloading binary: ${BINARY_URL}\n"
    curl --location --progress-bar "${BINARY_URL}" --output rill_${PLATFORM}.zip

    printf "\nDownloading checksum: ${CHECKSUM_URL}\n"
    curl --location --progress-bar "${CHECKSUM_URL}" --output checksums.txt

    printf "\nVerifying the SHA256 checksum of the downloaded binary:\n"
    shasum --algorithm 256 --ignore-missing --check checksums.txt

    printf "\nUnpacking rill_${PLATFORM}.zip\n"
    unzip -q rill_${PLATFORM}.zip
}

# Install the binary and ask for elevated permissions if needed
installBinary() {
    INSTALL_DIR="/usr/local/bin"
    if [ -w "${INSTALL_DIR}" ]; then
        printf "\nInstalling the Rill binary to: ${INSTALL_DIR}/rill\n"
        install rill "${INSTALL_DIR}/"
    else
        printf "\nElevated permissions required to install the Rill binary to: ${INSTALL_DIR}/rill\n"
        sudo install rill "${INSTALL_DIR}/"
    fi
}

# Run the installed binary and print the version
testInstalledBinary() {
    RILL_VERSION=$(rill version)
    printf "\nInstallation of ${RILL_VERSION} completed!\n"
    printf "\nThis application is extremely alpha and we want to hear from you if you have any questions or ideas to share! You can reach us in our Rill Discord server at https://rilldata.link/cli.\n"
}

# Parse input flag
case $1 in
    --nightly) VERSION=nightly;;
    --version) VERSION=${2:-latest};;
    *) VERSION=latest;;
esac

checkDependency curl
checkDependency shasum
checkDependency unzip
initPlatform
initTmpDir
downloadBinary
installBinary
testInstalledBinary
