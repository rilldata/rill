<script lang="ts">
  import YAMLEditor from "@rilldata/web-common/components/editor/YAMLEditor.svelte";
  import { indentGuide } from "@rilldata/web-common/components/editor/indent-guide";
  import { createLineStatusSystem } from "@rilldata/web-common/components/editor/plugins/line-status-decoration";
  import { editorTheme } from "@rilldata/web-common/components/editor/theme";
  import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";
  import { fileArtifactsStore } from "@rilldata/web-common/features/entity-management/file-artifacts-store";
  import { LIST_SLIDE_DURATION } from "@rilldata/web-common/layout/config";
  import { slide } from "svelte/transition";
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

  const { createUpdater, extension: lineStatusExtensions } =
    createLineStatusSystem();
  $: updateLineStatus = createUpdater([...mappedErrors, ...mappedSyntaxErrors]);

  let cursor;
  let hasFocus;
</script>

<div class="grid w-full" style:grid-template-rows="auto max-content">
  <div
    class="flex overflow-y-auto bg-white rounded-t border"
    class:rounded-b={!mappedErrors?.length && yaml?.length}
  >
    <div
      class="rounded border w-full"
      class:border-b-transparent={mainError && yaml?.length}
    >
      <YAMLEditor
        content={yaml}
        on:update
        on:cursor={(event) => {
          cursor = event.detail;
        }}
        bind:hasFocus
        plugins={[
          editorTheme(),
          placeholder,
          lineStatusExtensions,
          indentGuide,
        ]}
        stateFieldUpdaters={[updateLineStatus]}
      />
    </div>
  </div>
  {#if mainError && yaml?.length}
    <div
      transition:slide|local={{ duration: LIST_SLIDE_DURATION }}
      class="ui-editor-text-error ui-editor-bg-error border border-red-500 border-l-4 px-2 py-5"
    >
      <div class="flex gap-x-2 items-center">
        <CancelCircle />{mainError.message}
      </div>
    </div>
  {/if}
</div>
