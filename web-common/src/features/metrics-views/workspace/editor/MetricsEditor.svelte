<script lang="ts">
  import type { EditorView } from "@codemirror/basic-setup";
  import YAMLEditor from "@rilldata/web-common/components/editor/YAMLEditor.svelte";
  import { indentGuide } from "@rilldata/web-common/components/editor/indent-guide";
  import { createLineStatusSystem } from "@rilldata/web-common/components/editor/line-status";
  import type { LineStatus } from "@rilldata/web-common/components/editor/line-status/state";
  import { editorTheme } from "@rilldata/web-common/components/editor/theme";
  import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";
  import { LIST_SLIDE_DURATION } from "@rilldata/web-common/layout/config";
  import { slide } from "svelte/transition";
  import {
    createPlaceholder,
    createPlaceholderElement,
  } from "./create-placeholder";

  export let yaml: string;
  export let metricsDefName: string;
  export let view: EditorView;

  /** the main error to display on the bottom */
  export let error: LineStatus;

  /** create the line status system */
  const { extension: lineStatusExtensions } = createLineStatusSystem();

  let cursor;

  /** note: this codemirror plugin does actually utilize tanstack query, and the
   * instantiation of the underlying svelte component that defines the placeholder
   * must be instantiated in the component.
   */
  const placeholderElement = createPlaceholderElement(metricsDefName);
  const placeholder = createPlaceholder(placeholderElement.DOMElement);
</script>

<div
  class="editor pane flex flex-col w-full h-full content-stretch"
  style:height="calc(100vh - var(--header-height))"
>
  <div class="grow flex bg-white overflow-y-auto">
    <div
      class="border-white w-full overflow-y-auto"
      class:border-b-hidden={error && yaml?.length}
      class:border-red-500={error && yaml?.length}
    >
      <YAMLEditor
        content={yaml}
        bind:view
        on:update
        on:cursor={(event) => {
          cursor = event.detail;
        }}
        extensions={[
          editorTheme(),
          placeholder,
          lineStatusExtensions,
          indentGuide,
        ]}
      />
    </div>
  </div>
  {#if error && yaml?.length}
    <div
      transition:slide|local={{ duration: LIST_SLIDE_DURATION }}
      class="ui-editor-text-error ui-editor-bg-error border border-red-500 border-l-4 px-2 py-5"
    >
      <div class="flex gap-x-2 items-center">
        <CancelCircle />{error.message}
      </div>
    </div>
  {/if}
</div>
