<script lang="ts">
  import type { EditorView } from "@codemirror/view";
  import YAMLEditor from "@rilldata/web-common/components/editor/YAMLEditor.svelte";
  import { setLineStatuses } from "../../../components/editor/line-status";
  import { mapRuntimeErrorsToLines } from "../../metrics-views/errors";
  import { useSourceStore } from "../sources-store";

  export let yaml: string;
  export let sourceName: string;

  let editor: YAMLEditor;
  let view: EditorView;

  const sourceStore = useSourceStore();

  // PLACEDHOLDER
  /** note: this codemirror plugin does actually utilize tanstack query, and the
   * instantiation of the underlying svelte component that defines the placeholder
   * must be instantiated in the component.
   */
  // const placeholderElement = createPlaceholderElement(sourceName);
  // const placeholder = createPlaceholder(placeholderElement.DOMElement);

  // ERRORS
  // TODO: do the Source equivalent... if there's line-based errors
  // $: runtimeErrors = getMetricsDefErrors($fileArtifactsStore, metricsDefName);
  let runtimeErrors = [];
  $: lineBasedRuntimeErrors = mapRuntimeErrorsToLines(runtimeErrors, yaml);
  /** We display the mainError even if there are multiple errors elsewhere. */
  /** display the main error (the first in this array) at the bottom */
  $: mainError = [...lineBasedRuntimeErrors, ...(runtimeErrors || [])]?.at(0);
  /** If the errors change, run the following transaction.
   * Given that we are debouncing the core edit,
   */
  $: if (view) setLineStatuses(lineBasedRuntimeErrors, view);
</script>

<div class="editor flex flex-col border border-gray-200 rounded">
  <div class="grow flex bg-white overflow-y-auto rounded">
    <YAMLEditor
      bind:this={editor}
      bind:view
      content={yaml}
      on:update={(e) => sourceStore.set({ clientYAML: e.detail.content })}
    />
  </div>
</div>
