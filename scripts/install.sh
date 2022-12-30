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

# Check for dependent commands
checkCommand() {
    if ! command -v $1 &> /dev/null
    then
        echo "'$1' command not be found. This scripts depends on this command. Please install and retry."
        exit
    fi
}

# Download the binary and check the integrity using the SHA256 checksum
downloadBinary() {
    CDN="cdn.rilldata.com"
    BINARY_URL="https://${CDN}/rill/${VERSION}/rill_${PLATFORM}.zip"
    CHECKSUM_URL="https://${CDN}/rill/${VERSION}/checksums.txt"

    printf "Downloading binary: ${BINARY_URL}\n"
    curl --location --progress-bar "${BINARY_URL}" --output rill_${PLATFORM}.zip

    printf "\nDownloading checksum: ${CHECKSUM_URL}\n"
    curl --location --progress-bar "${CHECKSUM_URL}" --output checksums.txt

    printf "\nVerifying the SHA256 checksum of the downloaded binary:\n"
    shasum --algorithm 256 --ignore-missing --check checksums.txt

    printf "\nUnpacking rill_${PLATFORM}.zip\n"
    unzip rill_${PLATFORM}.zip
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
    *) VERSION=latest;;
esac

checkCommand shasum
checkCommand curl
checkCommand unzip
initPlatform
initTmpDir
downloadBinary
installBinary
testInstalledBinary
