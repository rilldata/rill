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

# Verify that a binary is available for the current combination of 'PLATFORM' and 'VERSION'
verifyAvailability() {
    if [ $VERSION == nightly ] && [ $PLATFORM == macos-arm64 ]; then
        printf "\nNightly builds are currently not published for ${PLATFORM}.\n\n"
        read -p "Do you want to install the nightly macos-x64 build which can run using Rosetta 2 instead? [y/N]" -n 1 -r -s < /dev/tty
        printf "\n\n"
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 0
        fi
        PLATFORM=macos-x64
    fi
}

# Create a temporary directory and setup deletion on script exit using the 'EXIT' signal
initTmpDir() {
    TMP_DIR=$(mktemp -d)
    trap 'rm -rf -- "${TMP_DIR}"' EXIT
    cd $TMP_DIR
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
    --nightly) VERSION=nightly/dist;;
    *) VERSION=latest;;
esac

initPlatform
verifyAvailability
initTmpDir
downloadBinary
installBinary
testInstalledBinary
