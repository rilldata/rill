<script lang="ts">
  import YAMLEditor from "@rilldata/web-common/components/editor/YAMLEditor.svelte";
  import { indentGuide } from "@rilldata/web-common/components/editor/indent-guide";
  import { createLineStatusSystem } from "@rilldata/web-common/components/editor/plugins/line-status-decoration";
  import { editorTheme } from "@rilldata/web-common/components/editor/theme";
  import { fileArtifactsStore } from "@rilldata/web-common/features/entity-management/file-artifacts-store";
  import { parseDocument } from "yaml";
  import {
    createPlaceholderElement,
    rillEditorPlaceholder,
  } from "../rill-editor-placeholder";

  export let yaml;
  export let metricsDefName;

  const placeholderSet = createPlaceholderElement(yaml);
  const placeholder = rillEditorPlaceholder(placeholderSet.DOMElement);
  $: placeholderSet.set(yaml);
  //placeholderSet.
  placeholderSet.on("test", (event) => {
    console.log(event.detail);
  });

  /** a temporary set of enums that shoul be emitted by orval's codegen */
  enum ConfigErrors {
    SourceNotSelected = "metrics view source not selected",
    SourceNotFound = "metrics view source not found",
    TimestampNotSelected = "metrics view timestamp not selected",
    TimestampNotFound = "metrics view selected timestamp not found",
    MissingDimension = "at least one dimension should be present",
    MissingMeasure = "at least one measure should be present",
    Malformed = "did not find expected key",
  }

  /** fixme: move to a file */
  enum YAMLSyntaxErrors {
    ALIAS_PROPS = "Alias node should not have any properties",
    BAD_ALIAS = "Alias node should be followed by a single non-empty plain scalar",
    BAD_DIRECTIVE = 'Expected "#", "YAML", "TAG" or whitespace but "%c" found',
    BAD_DQ_ESCAPE = 'Unexpected escape sequence "\\"%c"',
    BAD_INDENT = "Incorrect indentation in flow collection",
    BAD_LITERAL = "Unexpected end of the document within a single quoted scalar",
    BAD_PROP_ORDER = "Anchors and tags must be placed after the ?, : and - indicators",
    BAD_SCALAR_START = "Plain scalars cannot start with a block scalar indicator, or one of the two reserved characters: @ and `. To fix, use a block or quoted scalar for the value.",
    BLOCK_AS_IMPLICIT_KEY = "There's probably something wrong with the indentation, or you're trying to parse something like a: b: c, where it's not clear what's the key and what's the value.",
    BLOCK_IN_FLOW = "YAML scalars and collections both have block and flow styles. Flow is allowed within block, but not the other way around.",
    DUPLICATE_KEY = "Map keys must be unique",
    IMPOSSIBLE = "This really should not happen. If you encounter this error code, please file a bug.",
    KEY_OVER_1024_CHARS = "Keys longer than 1024 characters are not supported",
    MISSING_ANCHOR = "Alias node should be preceded by a non-empty anchor",
    MISSING_CHAR = "Some character or characters are missing here",
    MULTILINE_IMPLICIT_KEY = "Implicit keys need to be on a single line. Does the input include a plain scalar with a : followed by whitespace, which is getting parsed as a map key?",
    MULTIPLE_ANCHORS = "A node is only allowed to have one anchor.",
    MULTIPLE_DOCS = "A YAML stream may include multiple documents.",
    MULTIPLE_TAGS = "A node is only allowed to have one tag.",
    TAB_AS_INDENT = "Tabs are not allowed as indentation characters. Please use spaces instead.",
    TAG_RESOLVE_FAILED = "Failed to resolve tag",
    UNEXPECTED_TOKEN = "A token was encountered in a place where it wasn't expected.",
  }

  function runtimeErrorToLine(message: string, yaml: string) {
    const lines = yaml.split("\n");
    if (message === ConfigErrors.SourceNotFound) {
      /** if this is undefined, then the field isn't here either. */
      const line = lines.findIndex((line) => line.startsWith("model:"));
      return { line: line + 1, end: line, message, level: "error" };
    }
    if (message === ConfigErrors.TimestampNotFound) {
      const line =
        lines.findIndex((line) => line.startsWith("timeseries:")) + 1;
      return { line: line + 1, end: line, message, level: "error" };
    }
    if (message === ConfigErrors.MissingMeasure) {
      const line = lines.findIndex((line) => line.startsWith("measures:"));
      return { line: line + 1, end: line, message, level: "error" };
    }
    if (message === ConfigErrors.MissingDimension) {
      const line = lines.findIndex((line) => line.startsWith("dimensions:"));
      return { line: line + 1, end: line, message, level: "error" };
    }
    return { line: null, end: null, message, level: "error" };
  }

  function mapRuntimeErrorsToLines(errors) {
    if (!errors) return [];
    return errors
      .map((error) => {
        return runtimeErrorToLine(error.message, yaml);
      })
      .filter((error) => error.message !== ConfigErrors.Malformed);
  }

  $: path = Object.keys($fileArtifactsStore?.entities)?.find((key) => {
    return key.endsWith(`${metricsDefName}.yaml`);
  });
  let parsedYAML;
  $: if (yaml) parsedYAML = parseDocument(yaml);

  $: errors = $fileArtifactsStore?.entities?.[path]?.errors;
  $: mappedErrors = mapRuntimeErrorsToLines(errors);
  $: nonLineErrors = mappedErrors.filter((error) => !error.line);

  let mappedSyntaxErrors = [];
  $: if (parsedYAML?.errors?.length) {
    // parse the document and get errors.
    const parsedYAML = parseDocument(yaml);
    const syntaxErrors = parsedYAML.errors;
    mappedSyntaxErrors = syntaxErrors.map((error) => {
      return {
        line: error.linePos[0].line,
        message: error.message,
        level: "error",
      };
    });
  } else {
    mappedSyntaxErrors = [];
  }

  const { createUpdater, extension: lineStatusExtensions } =
    createLineStatusSystem();
  $: updateLineStatus = createUpdater([...mappedErrors, ...mappedSyntaxErrors]);

  let cursor;

  let hasFocus;
</script>

<div
  class="overflow-y-auto bg-white rounded {hasFocus ? 'ring-[1px]' : ''}"
  class:ring={hasFocus}
  class:ring-gray-300={hasFocus}
>
  <div class="rounded border">
    <YAMLEditor
      content={yaml}
      on:update
      on:cursor={(event) => {
        cursor = event.detail;
      }}
      bind:hasFocus
      plugins={[editorTheme(), placeholder, lineStatusExtensions, indentGuide]}
      stateFieldUpdaters={[updateLineStatus]}
    />
  </div>
</div>
