# Rill MCP Extensions

## Gemini Extension

### Documentation

Gemini CLI documentation can be found [here](https://github.com/google-gemini/gemini-cli/blob/main/docs/extension.md).

### Link a local extension

The `gemini extensions link` command will create a symbolic link from the extension installation directory to the development path.

This is useful so you don't have to run `gemini extensions update` every time you make changes you'd like to test.

```sh
gemini extensions link path/to/directory
```

## How it works

On startup, Gemini CLI looks for extensions in `<home>/.gemini/extensions`

Extensions exist as a directory that contains a `gemini-extension.json` file. For example:

`<home>/.gemini/extensions/my-extension/gemini-extension.json`
