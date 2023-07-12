# scripts/devtool

## Usage

1. To start all services from project's root directory, run the following command: `./scripts/devtool.sh`.
2. To exclude a specific service (admin/runtime/ui), add a flag with the service name set to false. For example, `./scripts/devtool.sh -ui=false` (NOTE: single hyphen) will only start admin and runtime.
3. To reset admin db use the `-reset` flag. Eg : `./scripts/devtool.sh -reset=true`