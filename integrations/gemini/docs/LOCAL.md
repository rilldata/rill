## Local Development

## Link a Local Extension

```bash
# Clone the Gemini extension repository
git clone https://github.com/rilldata/gemini.git && cd gemini
# Install dependencies
npm install
```

The `gemini extensions link` command creates a symbolic link from the extension installation directory to your development path. This is useful for testing changes without running `gemini extensions update` repeatedly.

```bash
gemini extensions link $(pwd)
```

### How It Works

On startup, Gemini CLI looks for extensions in `<home>/.gemini/extensions`

Extensions exist as a directory containing a `gemini-extension.json` file. For example:

```
<home>/.gemini/extensions/my-extension/gemini-extension.json
```

## Documentation

- [Gemini CLI Extension Documentation](https://github.com/google-gemini/gemini-cli/blob/main/docs/extension.md)
- [Google Cloud Configuration Guide](GEMINI.md)
