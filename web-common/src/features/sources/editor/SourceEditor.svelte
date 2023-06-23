<script lang="ts">
  import { indentGuide } from "@rilldata/web-common/components/editor/indent-guide";
  import { createLineStatusSystem } from "@rilldata/web-common/components/editor/line-status";
  import { editorTheme } from "@rilldata/web-common/components/editor/theme";
  import YAMLEditor from "@rilldata/web-common/components/editor/YAMLEditor.svelte";
  import { fileArtifactsStore } from "@rilldata/web-common/features/entity-management/file-artifacts-store";
  import { createEventDispatcher } from "svelte";
  import { parseDocument } from "yaml";

  export let yaml: string;
  export let sourceName: string;

  const dispatch = createEventDispatcher();

  /** a temporary set of enums that shoul be emitted by orval's codegen */
  enum ConfigErrors {
    SourceNotSelected = "metrics view source not selected",
    SourceNotFound = "metrics view source not found",
    SouceNotSelected = "metrics view source not selected",
    TimestampNotSelected = "metrics view timestamp not selected",
    TimestampNotFound = "metrics view selected timestamp not found",
    MissingDimension = "at least one dimension should be present",
    MissingMeasure = "at least one measure should be present",
    Malformed = "did not find expected key",
    InvalidTimeGrainForSmallest = "invalid time grain",
  }

  function runtimeErrorToLine(message: string, yaml: string) {
    const lines = yaml.split("\n");
    if (message === ConfigErrors.SouceNotSelected) {
      /** if this is undefined, then the field isn't here either. */
      const line = lines.findIndex((line) => line.startsWith("model: "));
      return { line: line + 1, end: line, message, level: "error" };
    }
    if (message.startsWith(ConfigErrors.InvalidTimeGrainForSmallest)) {
      const line = lines.findIndex((line) =>
        line.startsWith("smallest_time_grain:")
      );
      return { line: line + 1, end: line, message, level: "error" };
    }
    if (message === ConfigErrors.TimestampNotFound) {
      const line =
        lines.findIndex((line) => line.startsWith("timeseries:")) + 1;
      return { line: line, end: line, message, level: "error" };
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
    return key.endsWith(`${sourceName}.yaml`);
  });
  let parsedYAML;
  $: if (yaml) parsedYAML = parseDocument(yaml);

  $: errors = $fileArtifactsStore?.entities?.[path]?.errors;
  $: mappedErrors = mapRuntimeErrorsToLines(errors);

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

  /** We display the mainError even if there are multiple errors elsewhere. */
  $: mainError = [
    ...mappedSyntaxErrors,
    ...(mappedErrors || []),
    ...(errors || []),
  ]?.at(0);
  $: dispatch("error", mainError);

  /** create the line status system */
  const { createUpdater, extension: lineStatusExtensions } =
    createLineStatusSystem();
  $: updateLineStatus = createUpdater([...mappedErrors, ...mappedSyntaxErrors]);

  let cursor;

  /** note: this codemirror plugin does actually utilize tanstack query, and the
   * instantiation of the underlying svelte component that defines the placeholder
   * must be instantiated in the component.
   */
  // const placeholderElement = createPlaceholderElement(sourceName);
  // const placeholder = createPlaceholder(placeholderElement.DOMElement);
</script>

<div class="editor flex flex-col">
  <div class="grow flex bg-white overflow-y-auto">
    <div
      class="border-white w-full overflow-y-auto"
      class:border-b-hidden={mainError && yaml?.length}
      class:border-red-500={mainError && yaml?.length}
    >
      <YAMLEditor
        content={yaml}
        on:update
        on:cursor={(event) => {
          cursor = event.detail;
        }}
        extensions={[
          editorTheme(),
          // placeholder,
          lineStatusExtensions,
          indentGuide,
        ]}
        stateFieldUpdaters={[updateLineStatus]}
      />
    </div>
  </div>
</div>
