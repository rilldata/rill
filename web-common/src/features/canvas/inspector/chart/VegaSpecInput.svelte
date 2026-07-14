<script lang="ts">
  import { json } from "@codemirror/lang-json";
  import { EditorState } from "@codemirror/state";
  import { EditorView, placeholder } from "@codemirror/view";
  import { base as baseExtensions } from "@rilldata/web-common/components/editor/presets/base";

  export let value: string;
  export let onChange: (updatedSpec: string) => void;

  let error: string | null = null;

  const placeholderSpec = `Your Vega-Lite spec should look like this:
{
  "$schema": "https://vega.github.io/schema/vega-lite/v5.json",
  "data": {"name": "query1"},
  "mark": "bar",
  "encoding": {
    "x": {"field": "dimension", "type": "nominal"},
    "y": {"field": "measure", "type": "quantitative"}
  }
}

Data comes from your Metrics SQL queries. The first query is
available as {"name": "query1"}, the second as {"name": "query2"},
and so on.`;

  function initEditor(node: HTMLElement, initialValue: string) {
    // Guard against the feedback loop: when we dispatch a programmatic change
    // into the editor, the updateListener fires and would call onChange, which
    // would update the parent store, which would call update() again, ad infinitum.
    let externalUpdate = false;

    const editor = new EditorView({
      state: EditorState.create({
        doc: initialValue || "",
        extensions: [
          baseExtensions(),
          json(),
          placeholder(placeholderSpec),
          EditorView.updateListener.of((update) => {
            if (update.docChanged) {
              if (externalUpdate) return;

              const newValue = update.state.doc.toString();

              if (!newValue) {
                error = null;
                onChange("");
                return;
              }

              try {
                JSON.parse(newValue);
                error = null;
                onChange(newValue);
              } catch {
                error = "Invalid JSON";
              }
            }

            // Reformat on blur only, to avoid cursor jumps during typing
            if (update.focusChanged && !update.view.hasFocus) {
              if (externalUpdate) return;
              const newValue = update.state.doc.toString();
              if (!newValue) return;
              try {
                const formatted = JSON.stringify(JSON.parse(newValue), null, 2);
                if (formatted !== newValue) {
                  update.view.dispatch({
                    changes: {
                      from: 0,
                      to: update.state.doc.length,
                      insert: formatted,
                    },
                  });
                  onChange(formatted);
                }
              } catch {
                // already flagged as invalid above
              }
            }
          }),
          EditorView.theme({
            "&": { height: "500px" },
            ".cm-scroller": { overflow: "auto" },
          }),
        ],
      }),
      parent: node,
    });

    return {
      // Called by Svelte whenever the `value` prop changes.
      update(newValue: string) {
        const current = editor.state.doc.toString();
        if (current !== (newValue ?? "")) {
          externalUpdate = true;
          editor.dispatch({
            changes: {
              from: 0,
              to: editor.state.doc.length,
              insert: newValue ?? "",
            },
          });
          externalUpdate = false;
        }
      },
      destroy() {
        editor.destroy();
      },
    };
  }
</script>

<div>
  <div class="spec-editor-container" use:initEditor={value}></div>

  {#if error}
    <div class="text-red-500 text-sm px-3">{error}</div>
  {/if}
</div>

<style lang="postcss">
  .spec-editor-container {
    @apply my-2 pl-2 border-b border-gray-300;
  }

  :global(.spec-editor-container .cm-editor) {
    height: 400px;
    min-height: 100px;
    resize: vertical;
    overflow: hidden;
  }

  :global(.spec-editor-container .cm-editor .cm-scroller) {
    overflow: auto;
  }

  :global(.spec-editor-container .cm-gutter.cm-line-status-gutter) {
    display: none !important;
  }
</style>
