#!/usr/bin/env sh

# Determine the platform using 'OS' and 'ARCH'
initPlatform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    if [ "$OS" = "darwin" ] && [ "$ARCH" = "arm64" ]; then
        PLATFORM="darwin_arm64"
    elif [ "$OS" = "darwin" ] && [ "$ARCH" = "x86_64" ]; then
        PLATFORM="darwin_amd64"
    elif [ "$OS" = "linux" ] && [ "$ARCH" = "x86_64" ]; then
        PLATFORM="linux_amd64"
    elif [ "$OS" = "linux" ] && { [ "$ARCH" = "arm64" ] || [ "$ARCH" = "aarch64" ]; }; then
        PLATFORM="linux_arm64"
    else
        printf "Platform not supported: os=%s arch=%s\n" "$OS" "$ARCH"
        exit 1
    fi
}

# Create a temporary directory and setup deletion on script exit using the 'EXIT' signal
initTmpDir() {
    TMP_DIR=$(mktemp -d)
    trap 'rm -rf -- "${TMP_DIR}"' EXIT
    cd "$TMP_DIR"
}

# Ensure that dependency is installed and executable, exit and print help message if not
checkDependency() {
    if ! [ -x "$(command -v "$1")" ]; then
        printf "'%s' could not be found, this script depends on it, please install and try again.\n" "$1"
        exit 1
    fi
}

# Ensure that 'git' is installed and executable, exit and print help message if not
checkGitDependency() {
    if ! [ -x "$(command -v git)" ]; then
        publishSyftEvent git_missing
        printf "Git could not be found, Rill depends on it, please install and try again.\n\n"
        printf "Helpful instructions: https://github.com/git-guides/install-git\n"
        exit 1
    fi
}

# Ensure that either 'shasum' or 'sha256sum' is installed and executable, exit and print help message if not
resolveShasumDependency() {
    if [ -x "$(command -v shasum)" ]; then
        sha256_verify="shasum --algorithm 256 --ignore-missing --check"
    elif [ -x "$(command -v sha256sum)" ]; then
        sha256_verify="sha256sum --ignore-missing --check"
    else
        printf "neither 'shasum' or 'sha256sum' could be found, this script depends on one of them, please install one of them and try again.\n"
        exit 1
    fi
}

# Download the binary and check the integrity using the SHA256 checksum
downloadBinary() {
    CDN="cdn.rilldata.com"

    LATEST_URL="https://${CDN}/rill/latest.txt"
    if [ "${VERSION}" = "latest" ]; then
        VERSION=$(curl --silent --show-error ${LATEST_URL})
    fi
    BINARY_URL="https://${CDN}/rill/${VERSION}/rill_${PLATFORM}.zip"
    CHECKSUM_URL="https://${CDN}/rill/${VERSION}/checksums.txt"

    if [ "$NON_INTERACTIVE" = "true" ]; then
        set -- "--silent" "--show-error"
    else
        set -- "--progress-bar"
    fi

    printf "Downloading binary: %s\n" "$BINARY_URL"
    curl --location "$@" "${BINARY_URL}" --output rill_${PLATFORM}.zip

    printf "\nDownloading checksum: %s\n" "$CHECKSUM_URL"
    curl --location "$@" "${CHECKSUM_URL}" --output checksums.txt

    printf "\nVerifying the SHA256 checksum of the downloaded binary:\n"
    ${sha256_verify} checksums.txt

    printf "\nUnpacking rill_%s.zip\n" "$PLATFORM"
    unzip -q rill_${PLATFORM}.zip
}

# Print install options
printInstallOptions() {
    printf "\nWhere would you like to install rill?  (Default [1])\n\n"
    printf "[1]  /usr/local/bin/rill  [recommended, but requires sudo privileges]\n"
    printf "[2]  ~/.rill/rill         [directory will be created & path configured]\n"
    printf "[3]  ./rill               [download to the current directory]\n\n"
}

# Ask for preferred install option
promptInstallChoice() {
    printf "Pick install option: (1/2/3)\n"
    read -r ans </dev/tty;
    case $ans in
        1|"")
            INSTALL_DIR="/usr/local/bin"
            ;;
        2)
            INSTALL_DIR="$HOME/.rill"
            ;;
        3)
            INSTALL_DIR=$(pwd)
            ;;
        *)
            printf "\nInvalid option '%s'\n\n" "$ans"
            promptInstallChoice
            ;;
    esac
}

# Check conflicting installation and exit with a help message
checkConflictingInstallation() {
    if [ -x "$(command -v rill)" ]; then
        INSTALLED_RILL="$(command -v rill)"
        if [ -x "$(command -v brew)" ] && brew list rilldata/tap/rill >/dev/null 2>&1; then
            printf "There is a conflicting version of Rill installed using Brew.\n\n"
            printf "To upgrade using Brew, run 'brew upgrade rilldata/tap/rill'.\n\n"
            printf "To use this script to install Rill, run 'brew remove rilldata/tap/rill' to remove the conflicting version and try again.\n"
            exit 1
        elif [ "$INSTALLED_RILL" != "${INSTALL_DIR}/rill" ]; then
            printf "There is a conflicting version of Rill installed at '%s'\n\n" "$INSTALLED_RILL"
            printf "To use this script to install Rill, remove the conflicting version and try again.\n"
            exit 1
        fi
    fi
}

# Install the binary and ask for elevated permissions if needed
installBinary() {
    if [ "$INSTALL_DIR" = "/usr/local/bin" ]; then
        printf "\nElevated permissions required to install the Rill binary to: %s/rill\n" "$INSTALL_DIR"
        sudo install -d "$INSTALL_DIR"
        sudo install rill "$INSTALL_DIR"
    else
        install -d "$INSTALL_DIR"
        install rill "$INSTALL_DIR"
    fi
    cd - > /dev/null
}

# Run the installed binary and print the version
testInstalledBinary() {
    RILL_VERSION=$("$INSTALL_DIR"/rill version)
    "$INSTALL_DIR"/rill verify-install 1>/dev/null || true
    printf "\nInstallation of %s completed!\n" "$RILL_VERSION"
}

# Print 'rill start' help intrcutions
printStartHelp() {
    boldon=$(tput smso)
    boldoff=$(tput rmso)

    if [ "$INSTALL_DIR" = "/usr/local/bin" ]; then
        printf "\nTo start a new project in Rill, execute the command:\n\n %srill start my-rill-project%s\n\n" "$boldon" "$boldoff"
    elif [ "$INSTALL_DIR" = "$HOME/.rill" ]; then
        printf "\nTo start a new project in Rill, open a %snew terminal%s and execute the command:\n\n %srill start my-rill-project%s\n\n" "$boldon" "$boldoff" "$boldon" "$boldoff"
    elif [ "$INSTALL_DIR" = "$(pwd)" ]; then
        printf "\nTo start a new project in Rill, execute the command:\n\n %s./rill start my-rill-project%s\n\n" "$boldon" "$boldoff"
    fi
}

# Publish Syft install telemetry event, can be disabled by setting the 'RILL_INSTALL_DISABLE_TELEMETRY' environment variable
publishSyftEvent() {
    SYFT_URL=https://event.syftdata.com/log
    SYFT_ID=clp76quhs0006l908bux79l4v
    if [ -z "$RILL_INSTALL_DISABLE_TELEMETRY" ]; then
        curl --silent --show-error --header "Authorization: ${SYFT_ID}" --header "Content-Type: application/json" --data "{\"event_name\":\"$1\"}" $SYFT_URL > /dev/null || true >&2
    fi
}

# Add the Rill binary to the PATH via configuration of the shells we detect on the system
addPathConfigEntries() {
    PATH_CONFIG_LINE="export PATH=\$HOME/.rill:\$PATH # Added by Rill install"

    if [ "$INSTALL_DIR" = "$HOME/.rill" ]; then
        for f in "$HOME/.bashrc" "$HOME/.zshrc"; do
            if [ -f "$f" ]; then
                if ! grep -Fxq "$PATH_CONFIG_LINE" "$f"; then
                    printf "\nWould you like to add 'rill' to your PATH by adding an entry in '%s'? (Y/n)\n" "$f"
                    read -r ans </dev/tty;
                    case $ans in
                        n)
                            ;;
                        *)
                            printf "\n%s\n" "$PATH_CONFIG_LINE" >> "$f"
                            ;;
                    esac
                fi
            fi
        done
    fi
}

# Remove PATH configurations, we have to do handle this slightly different based on OS because of platform variations in 'sed' behaviour
removePathConfigEntries() {
    for f in "$HOME/.bashrc" "$HOME/.zshrc"; do
        if [ -f "$f" ]; then
            if [ "$OS" = "darwin" ]; then
                sed -i "" -e '/# Added by Rill install/d' "$f"
            elif [ "$OS" = "linux" ]; then
                sed -i -e '/# Added by Rill install/d' "$f"
            fi
        fi
    done
}

# Check if the install directory (or its parent, if it doesn't exist yet) is writable
installDirIsWritable() {
    if [ -d "$INSTALL_DIR" ]; then
        [ -w "$INSTALL_DIR" ]
    else
        [ -w "$(dirname "$INSTALL_DIR")" ]
    fi
}

# Resolve the install directory
resolveInstallDir() {
    # Detect previous installation
    if [ -x "$(command -v rill)" ] && [ -z "${INSTALL_DIR}" ]; then
        INSTALLED_RILL="$(command -v rill)"
        if [ "$INSTALLED_RILL" = "/usr/local/bin/rill" ]; then
            INSTALL_DIR="/usr/local/bin"
        elif [ "$INSTALLED_RILL" = "$HOME/.rill/rill" ]; then
            INSTALL_DIR="$HOME/.rill"
        fi
    fi

    # Handle non-interactive scenarios where prompt or sudo are not possible
    if [ "$NON_INTERACTIVE" = "true" ]; then
        # If the install directory was explicitly set and requires sudo, we error
        if [ -n "$INSTALL_DIR" ] && [ "$INSTALL_DIR_EXPLICIT" = "true" ] && ! installDirIsWritable; then
            printf "Install directory '%s' requires elevated permissions, which are not available in non-interactive mode.\n" "$INSTALL_DIR"
            exit 1
        fi

        # If the install directory is not set, or the previous installation path requires sudo, we default to installing in the current directory
        if [ -z "$INSTALL_DIR" ]; then
            printf "Non-interactive shell detected; defaulting to install in current directory.\n"
            INSTALL_DIR=$(pwd)
        elif ! installDirIsWritable; then
            printf "Non-interactive shell detected; previous installation at '%s' requires elevated permissions; defaulting to install in current directory.\n" "$INSTALL_DIR"
            INSTALL_DIR=$(pwd)
        fi

        return
    fi

    # If there is a previous or explicit installation path, we're done
    if [ -n "${INSTALL_DIR}" ]; then
        return
    fi

    # Prompt for install directory if there is no previous installation and we are in an interactive shell
    printInstallOptions
    promptInstallChoice
    checkConflictingInstallation # Only check for conflicts in interactive, non-explicit scenarios
}

# Install Rill on the system
installRill() {
    publishSyftEvent install
    checkDependency curl
    checkDependency unzip
    checkGitDependency
    resolveShasumDependency
    initPlatform
    resolveInstallDir
    initTmpDir
    downloadBinary
    installBinary
    testInstalledBinary
    if [ "$NON_INTERACTIVE" != "true" ]; then
        addPathConfigEntries
    fi
    printStartHelp
    publishSyftEvent installed
}

# Uninstall Rill from the system, this function is aware of both the privileged and unprivileged install methods
uninstallRill() {
    checkDependency sed
    initPlatform

    if [ -f "/usr/local/bin/rill" ]
    then
        printf "\nElevated permissions required to uninstall the Rill binary from: '/usr/local/bin/rill'\n"
        sudo rm /usr/local/bin/rill
    fi

    rm -f "$HOME/.rill/rill"
    removePathConfigEntries

    printf "Uninstall of Rill completed\n"
}

set -e

# Default values
INSTALL_DIR_EXPLICIT=false
if ! [ -t 0 ]; then # Detect non-interactive environments (e.g. piped input, CI, subprocesses)
    # When invoked as a subprocess of the Rill CLI (e.g. `rill upgrade`), stay interactive.
    # Older versions of Rill don't pass stdin through, but the user is still at a terminal.
    PARENT_NAME=""
    if [ -f "/proc/$PPID/comm" ]; then
        PARENT_NAME=$(cat "/proc/$PPID/comm" 2>/dev/null)
    elif command -v ps >/dev/null 2>&1; then
        PARENT_NAME=$(ps -o comm= -p "$PPID" 2>/dev/null)
    fi
    if [ "$(basename "$PARENT_NAME" 2>/dev/null)" != "rill" ]; then
        NON_INTERACTIVE=${NON_INTERACTIVE:-true}
    fi
fi
NON_INTERACTIVE=${NON_INTERACTIVE:-false}

# Parse input flag
case $1 in
    --uninstall)
        uninstallRill
        ;;
    --nightly)
        VERSION=nightly
        installRill
        ;;
    --version)
        VERSION=${2:-latest}
        installRill
        ;;
    --non-interactive)
        if [ -n "$2" ]; then
            INSTALL_DIR="$2"
            INSTALL_DIR_EXPLICIT=true
        fi
        VERSION=${3:-latest}
        NON_INTERACTIVE=true
        installRill
        ;;
    *)
        VERSION=latest
        installRill
        ;;
esac
