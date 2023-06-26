<script lang="ts">
  import type { EditorView } from "@codemirror/basic-setup";
  import EditorContainer from "@rilldata/web-common/components/editor/EditorContainer.svelte";
  import YAMLEditor from "@rilldata/web-common/components/editor/YAMLEditor.svelte";
  import type { LineStatus } from "@rilldata/web-common/components/editor/line-status/state";
  import {
    createPlaceholder,
    createPlaceholderElement,
  } from "./create-placeholder";

  export let yaml: string;
  export let metricsDefName: string;
  export let view: EditorView;

  /** the main error to display on the bottom */
  export let error: LineStatus;

  let cursor;

  /** note: this codemirror plugin does actually utilize tanstack query, and the
   * instantiation of the underlying svelte component that defines the placeholder
   * must be instantiated in the component.
   */
  const placeholderElement = createPlaceholderElement(metricsDefName);
  const placeholder = createPlaceholder(placeholderElement.DOMElement);
</script>

<EditorContainer {error} hasContent={!!yaml?.length}>
  <YAMLEditor
    content={yaml}
    bind:view
    on:update
    on:cursor={(event) => {
      cursor = event.detail;
    }}
    extensions={[placeholder]}
  />
</EditorContainer>
