#!/usr/bin/env bash
set -e

CDN="cdn.rilldata.com"
INSTALL_DIR="/usr/local/bin"

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
    LATEST_URL="https://${CDN}/rill/latest.txt"
    if [ "${VERSION}" == "latest" ]; then
        VERSION=$(curl --silent --show-error ${LATEST_URL})
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

# Check conflicting installation and exit with a help message
checkConflictingInstallation() {
    if [ -x "$(command -v rill)" ]; then
        INSTALLED_RILL="$(command -v rill)"
        if [ -x "$(command -v brew)" ] && brew list rilldata/tap/rill &>/dev/null; then
            printf "There is a conflicting version of Rill installed using Brew.\n\n"
            printf "To upgrade using Brew, run 'brew upgrade rilldata/tap/rill'.\n\n"
            printf "To use this script to install Rill, run 'brew remove rilldata/tap/rill' to remove the conflicting version and try again.\n"
            exit 1
        elif [ $INSTALLED_RILL != "${INSTALL_DIR}/rill" ]; then
            printf "There is a conflicting version of Rill installed at '${INSTALLED_RILL}'\n\n"
            printf "To use this script to install Rill, remove the conflicting version and try again.\n"
            exit 1
        fi
    fi
}

# Ask for elevated permissions to install the binary
installBinary() {
    printf "\nElevated permissions required to install the Rill binary to: ${INSTALL_DIR}/rill\n"
    sudo install -d ${INSTALL_DIR}
    sudo install rill "${INSTALL_DIR}/"
}

# Run the installed binary and print the version
testInstalledBinary() {
    RILL_VERSION=$(rill version)
    rill verify-install 1>/dev/null || true
    boldon=`tput smso`
    boldoff=`tput rmso`
    printf "\nInstallation of ${RILL_VERSION} completed!\n"
    printf "\nTo start a new project in Rill, execute the command:\n\n ${boldon}rill start my-rill-project${boldoff}\n\n"
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
checkConflictingInstallation
initTmpDir
downloadBinary
installBinary
testInstalledBinary
