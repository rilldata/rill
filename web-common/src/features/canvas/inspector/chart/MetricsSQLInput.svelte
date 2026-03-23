<script lang="ts">
  import { sql } from "@codemirror/lang-sql";
  import { EditorState } from "@codemirror/state";
  import { EditorView, placeholder } from "@codemirror/view";
  import { base as baseExtensions } from "@rilldata/web-common/components/editor/presets/base";
  import { DuckDBSQL } from "@rilldata/web-common/components/editor/presets/duckDBDialect";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import { onDestroy, tick } from "svelte";

  export let key: string;
  export let label: string | undefined;
  export let description: string | undefined;
  export let value: string[] = [];
  export let onChange: (updatedSQLs: string[]) => void;

  // Ensure at least one editor is always shown
  if (!value || value.length === 0) value = [""];

  let editorContainers: HTMLElement[] = [];
  let editors: EditorView[] = [];

  function createEditor(container: HTMLElement, idx: number) {
    const editor = new EditorView({
      state: EditorState.create({
        doc: value[idx] || "",
        extensions: [
          baseExtensions(),
          sql({ dialect: DuckDBSQL }),
          placeholder("SELECT * FROM metrics"),
          EditorView.updateListener.of((update) => {
            if (update.docChanged) {
              updateSQL(idx, update.state.doc.toString());
            }
          }),
          EditorView.theme({
            "&": { maxHeight: "150px" },
            ".cm-scroller": { overflow: "auto" },
          }),
        ],
      }),
      parent: container,
    });
    editors[idx] = editor;
  }

  function initEditor(node: HTMLElement, idx: number) {
    editorContainers[idx] = node;
    createEditor(node, idx);

    return {
      destroy() {
        editors[idx]?.destroy();
      },
    };
  }

  function updateSQL(idx: number, newSQL: string) {
    const updated = value.slice();
    updated[idx] = newSQL;
    onChange(updated);
  }

  async function addQuery() {
    onChange([...value, ""]);
    await tick();
  }

  function removeQuery(idx: number) {
    if (value.length === 1) return;
    // Destroy all editors; they'll be re-created by the action directive
    editors.forEach((e) => e?.destroy());
    editors = [];
    const updated = value.slice();
    updated.splice(idx, 1);
    onChange(updated);
  }

  onDestroy(() => {
    editors.forEach((e) => e?.destroy());
  });
</script>

<div class="sql-input-container">
  {#if label}
    <InputLabel hint={description} small {label} id={key} />
  {/if}

  <div class="queries">
    {#each value as _sql, idx (idx)}
      <div class="query-block">
        {#if value.length > 1}
          <div class="query-header">
            <span class="query-number">Query {idx + 1}</span>
            <button
              class="remove-btn"
              on:click={() => removeQuery(idx)}
              aria-label="Remove query {idx + 1}"
            >
              <svg width="12" height="12" viewBox="0 0 16 16" fill="none">
                <path
                  d="M4 8h8"
                  stroke="currentColor"
                  stroke-width="1.5"
                  stroke-linecap="round"
                />
              </svg>
            </button>
          </div>
        {/if}
        <div class="editor-wrapper" use:initEditor={idx} />
      </div>
    {/each}
  </div>

  <button class="add-query-btn" on:click={addQuery}>
    <svg width="12" height="12" viewBox="0 0 16 16" fill="none">
      <path
        d="M8 3v10M3 8h10"
        stroke="currentColor"
        stroke-width="1.5"
        stroke-linecap="round"
      />
    </svg>
    Add query
  </button>
</div>

<style lang="postcss">
  .sql-input-container {
    @apply flex flex-col gap-2;
  }

  .queries {
    @apply flex flex-col gap-2;
  }

  .query-block {
    @apply flex flex-col gap-1;
  }

  .query-header {
    @apply flex items-center justify-between;
  }

  .query-number {
    @apply text-[10px] font-medium text-gray-400 uppercase tracking-wide;
  }

  .remove-btn {
    @apply p-0.5 rounded text-gray-300 bg-transparent border-none cursor-pointer;
    @apply opacity-0 transition-all duration-150;
  }

  .query-block:hover .remove-btn {
    @apply opacity-100;
  }

  .remove-btn:hover {
    @apply text-red-400 bg-red-50;
  }

  .editor-wrapper {
    @apply border border-gray-200 rounded-md overflow-hidden;
    @apply transition-colors duration-150;
  }

  .editor-wrapper:focus-within {
    @apply border-primary-400 ring-1 ring-primary-200;
  }

  :global(.editor-wrapper .cm-editor) {
    min-height: 48px;
    max-height: 150px;
  }

  :global(.editor-wrapper .cm-editor .cm-scroller) {
    overflow: auto;
  }

  :global(.editor-wrapper .cm-gutter.cm-line-status-gutter) {
    display: none !important;
  }

  .add-query-btn {
    @apply flex items-center gap-1.5 self-start;
    @apply px-0 py-0.5 text-[11px] text-gray-400;
    @apply bg-transparent border-none cursor-pointer;
    @apply transition-colors duration-150;
  }

  .add-query-btn:hover {
    @apply text-primary-500;
  }
</style>
