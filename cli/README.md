# `cli`

*NOTE: This folder is a work in progress. See `src/cli/` for the CLI currently used in Rill Developer.*

## Example usage:

1. In a terminal, Run following commands to use rill cli:
```
make cli
./rill --help 
```
2. To install it via homebrew
```
brew install rilldata/rill-developer/rill
rill --help
```

3. You can also try our example:
    1. List of available examples
    ```
    rill init list
    ```
    2. Initialize the example project, default to `default`, default directory `.`
    ```
    rill init --example <example project name> --dir <directory to migrate example project>
    ```
    3. Start rill with hydration of example project, `start --help` for other available flags
    ```
    rill start --dir <example project directory>
    ```
4. See our documentation for more information by running
```
rill docs
```

*NOTE: Few of the CLI commands are work in progress, it will just print the message eg. `command Name is called`
