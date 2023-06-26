# scripts/devtool

## Usage

1. Run following command to start all services from project's root directory `sh scripts/devtool/dev.sh`.
2. To not admin/runtime/ui pass a flag with the service name as false. `sh scripts/devtool/dev.sh -ui=false` (NOTE: single hyphen) will only start admin and runtime.
3. To reset admin db pass `-reset` flag. Eg : `sh scripts/devtool/dev.sh -reset`