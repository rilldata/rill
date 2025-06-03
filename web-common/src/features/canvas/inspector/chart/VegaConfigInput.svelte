<script lang="ts">
  import { json } from "@codemirror/lang-json";
  import { EditorState } from "@codemirror/state";
  import { EditorView, placeholder } from "@codemirror/view";
  import { base as baseExtensions } from "@rilldata/web-common/components/editor/presets/base";
  import { onDestroy, onMount } from "svelte";
  import { get } from "svelte/store";
  import type { BaseCanvasComponent } from "../../components/BaseCanvasComponent";

  export let component: BaseCanvasComponent;

  const KEY = "vl_config";
  let error: string | null = null;
  let configEditor: EditorView;
  let editorContainer: HTMLElement;

  const placeholderConfig = `Your config should look like this:
{
  "axisX": {
    "grid": false,
  },
  "range": {
    "category": [
      "#ff7f0e",
      "#2ca02c",
    ]
  }
  ...
}`;

  onMount(() => {
    const paramValues = get(component.specStore);
    configEditor = new EditorView({
      state: EditorState.create({
        doc: (paramValues[KEY] as string | undefined) || "",
        extensions: [
          baseExtensions(),
          json(),
          placeholder(placeholderConfig),
          EditorView.updateListener.of((update) => {
            if (update.docChanged) {
              const newValue = update.state.doc.toString();
              paramValues[KEY] = newValue;

              if (!newValue) {
                error = null;
                component.updateProperty(KEY, null);
                return;
              }

              try {
                const parsed = JSON.parse(newValue);
                const formatted = JSON.stringify(parsed, null, 2);
                error = null;
                component.updateProperty(KEY, formatted);
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
      configEditor.destroy();
    };
  });

  onDestroy(() => configEditor?.destroy());
</script>

<div>
  <div class="border-y px-3 py-5">
    Enter desired <a
      href="https://vega.github.io/vega-lite/docs/config.html"
      target="_blank"
      >Vega-Lite config
    </a>below
  </div>
  <div bind:this={editorContainer} class="config-editor-container" />

  {#if error}
    <div class="text-red-500 text-sm px-3">{error}</div>
  {/if}
</div>

<style lang="postcss">
  .config-editor-container {
    @apply my-2 pl-2 border-b border-gray-300;
  }

  :global(.config-editor-container .cm-editor) {
    height: 500px;
  }

  :global(.config-editor-container .cm-editor .cm-scroller) {
    overflow: auto;
  }

  :global(.config-editor-container .cm-gutter.cm-line-status-gutter) {
    display: none !important;
  }
</style>
