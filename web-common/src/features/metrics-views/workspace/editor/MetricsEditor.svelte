<script lang="ts">
  import type { EditorView } from "@codemirror/basic-setup";
  import EditorContainer from "@rilldata/web-common/components/editor/EditorContainer.svelte";
  import YAMLEditor from "@rilldata/web-common/components/editor/YAMLEditor.svelte";
  import type { LineStatus } from "@rilldata/web-common/components/editor/line-status/state";
  import type { V1ReconcileError } from "@rilldata/web-common/runtime-client";
  import {
    createPlaceholder,
    createPlaceholderElement,
  } from "./create-placeholder";

  export let yaml: string;
  export let metricsDefName: string;
  export let view: EditorView;

  /** the main error to display on the bottom */
  export let error: LineStatus | V1ReconcileError;

  /** note: this codemirror plugin does actually utilize tanstack query, and the
   * instantiation of the underlying svelte component that defines the placeholder
   * must be instantiated in the component.
   */
  const placeholderElement = createPlaceholderElement(metricsDefName);
  $: if (view) {
    placeholderElement.setEditorView(view);
  }

  const placeholder = createPlaceholder(placeholderElement.DOMElement);

  let editor: YAMLEditor;
</script>

<EditorContainer error={!!yaml?.length ? error : undefined}>
  <YAMLEditor
    bind:this={editor}
    content={yaml}
    bind:view
    on:update
    extensions={[placeholder]}
  />
</EditorContainer>
