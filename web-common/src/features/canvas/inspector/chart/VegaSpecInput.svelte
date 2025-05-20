<script lang="ts">
  import { json } from "@codemirror/lang-json";
  import { EditorState } from "@codemirror/state";
  import { EditorView, placeholder } from "@codemirror/view";
  import { base as baseExtensions } from "@rilldata/web-common/components/editor/presets/base";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import { onDestroy, onMount } from "svelte";

  export let key: string;
  export let label: string | undefined;
  export let description: string | undefined;
  export let value: string;
  export let onChange: (updatedSpec: string) => void;

  let error: string | null = null;
  let specEditor: EditorView;
  let editorContainer: HTMLElement;

  const placeholderSpec = `Your Vega-Lite spec should look like this:
{
  "$schema": "https://vega.github.io/schema/vega-lite/v5.json",
  "description": "A simple bar chart with embedded data.",
  "data": {
    "values": [
      {"a": "A", "b": 28},
      {"a": "B", "b": 55},
      {"a": "C", "b": 43}
    ]
  },
  "mark": "bar",
  "encoding": {
    "x": {"field": "a", "type": "nominal", "axis": {"labelAngle": 0}},
    "y": {"field": "b", "type": "quantitative"}
  }
}`;

  onMount(() => {
    specEditor = new EditorView({
      state: EditorState.create({
        doc: value || "",
        extensions: [
          baseExtensions(),
          json(),
          placeholder(placeholderSpec),
          EditorView.updateListener.of((update) => {
            if (update.docChanged) {
              const newValue = update.state.doc.toString();

              if (!newValue) {
                error = null;
                onChange("");
                return;
              }

              try {
                const parsed = JSON.parse(newValue);
                const formatted = JSON.stringify(parsed, null, 2);
                error = null;
                onChange(formatted);
              } catch {
                error = "Invalid JSON";
              }
            }
          }),
          EditorView.theme({
            "&": { height: "500px" },
            ".cm-scroller": { overflow: "auto" },
          }),
        ],
      }),
      parent: editorContainer,
    });

    return () => {
      specEditor.destroy();
    };
  });

  onDestroy(() => specEditor?.destroy());
</script>

<div>
  <InputLabel hint={description} small label={label ?? key} id={key} />
  <div bind:this={editorContainer} class="spec-editor-container" />

  {#if error}
    <div class="text-red-500 text-sm px-3">{error}</div>
  {/if}
</div>

<style lang="postcss">
  .spec-editor-container {
    @apply my-2 pl-2 border-b border-gray-300;
  }

  :global(.spec-editor-container .cm-editor) {
    height: 300px;
  }

  :global(.spec-editor-container .cm-editor .cm-scroller) {
    overflow: auto;
  }

  :global(.spec-editor-container .cm-gutter.cm-line-status-gutter) {
    display: none !important;
  }
</style>
