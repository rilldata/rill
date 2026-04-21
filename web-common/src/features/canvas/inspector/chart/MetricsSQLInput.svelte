<script lang="ts">
  import { sql } from "@codemirror/lang-sql";
  import { EditorState } from "@codemirror/state";
  import { EditorView, placeholder } from "@codemirror/view";
  import { base as baseExtensions } from "@rilldata/web-common/components/editor/presets/base";
  import { DuckDBSQL } from "@rilldata/web-common/components/editor/presets/duckDBDialect";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import Trash from "@rilldata/web-common/components/icons/Trash.svelte";
  import { tick } from "svelte";

  export let key: string;
  export let label: string | undefined;
  export let description: string | undefined;
  export let value: string[] = [];
  export let onChange: (updatedSQLs: string[]) => void;

  // Each entry is a stable { id, sql } pair so Svelte keys by id, not index.
  type Entry = { id: number; sql: string };
  let nextId = 0;

  function toEntries(sqls: string[]): Entry[] {
    return sqls.map((s) => ({ id: nextId++, sql: s }));
  }

  let entries: Entry[] = toEntries(!value || value.length === 0 ? [""] : value);

  // When the value prop changes externally (e.g. AI agent writes to YAML), sync
  // entries. The use:initEditor action's update() hook then syncs each editor's
  // content when its entry changes, so no manual EditorView tracking is needed.
  $: {
    const incoming = !value || value.length === 0 ? [""] : value;
    if (incoming.length !== entries.length) {
      // Query count changed — rebuild entries so Svelte re-keys the #each block,
      // destroying old editors and creating fresh ones with the new content.
      entries = toEntries(incoming);
    } else {
      // Same count — update sql on each entry; Svelte passes the new entry to
      // the action's update() which dispatches the change into the editor.
      entries = entries.map((e, i) => ({ ...e, sql: incoming[i] }));
    }
  }

  function initEditor(node: HTMLElement, entry: Entry) {
    // Prevent the feedback loop: programmatic dispatches must not trigger onChange.
    let externalUpdate = false;

    const editor = new EditorView({
      state: EditorState.create({
        doc: entry.sql,
        extensions: [
          baseExtensions(),
          sql({ dialect: DuckDBSQL }),
          placeholder("SELECT * FROM metrics"),
          EditorView.updateListener.of((update) => {
            if (update.docChanged && !externalUpdate) {
              updateSQL(entry.id, update.state.doc.toString());
            }
          }),
          EditorView.theme({
            "&": { height: "150px" },
            ".cm-scroller": { overflow: "auto" },
          }),
        ],
      }),
      parent: node,
    });

    return {
      // Called by Svelte whenever the entry prop passed to use:initEditor changes.
      update(newEntry: Entry) {
        const current = editor.state.doc.toString();
        if (current !== newEntry.sql) {
          externalUpdate = true;
          editor.dispatch({
            changes: {
              from: 0,
              to: editor.state.doc.length,
              insert: newEntry.sql,
            },
          });
          externalUpdate = false;
        }
        // Keep the closure's entry.id in sync so updateSQL references the right id.
        entry = newEntry;
      },
      destroy() {
        editor.destroy();
      },
    };
  }

  function updateSQL(id: number, newSQL: string) {
    entries = entries.map((e) => (e.id === id ? { ...e, sql: newSQL } : e));
    onChange(entries.map((e) => e.sql));
  }

  async function addQuery() {
    entries = [...entries, { id: nextId++, sql: "" }];
    onChange(entries.map((e) => e.sql));
    await tick();
  }

  function removeQuery(id: number) {
    if (entries.length === 1) return;
    entries = entries.filter((e) => e.id !== id);
    onChange(entries.map((e) => e.sql));
  }
</script>

<div class="sql-input-container">
  {#if label}
    <InputLabel hint={description} small {label} id={key} />
  {/if}

  <div class="queries">
    {#each entries as entry, idx (entry.id)}
      <div class="query-block">
        {#if entries.length > 1}
          <div class="query-header">
            <span class="query-number">Query {idx + 1}</span>
            <button
              class="remove-btn"
              onclick={() => removeQuery(entry.id)}
              aria-label="Remove query {idx + 1}"
            >
              <Trash size="14px" />
            </button>
          </div>
        {/if}
        <div class="editor-wrapper" use:initEditor={entry}></div>
      </div>
    {/each}
  </div>

  <button class="add-query-btn" onclick={addQuery}>
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
    @apply p-1 rounded text-gray-400 bg-transparent border-none cursor-pointer;
    @apply transition-colors duration-150;
  }

  .remove-btn:hover {
    @apply text-red-500 bg-red-50;
  }

  .editor-wrapper {
    @apply border border-gray-200 rounded-md overflow-hidden;
    @apply transition-colors duration-150;
  }

  .editor-wrapper:focus-within {
    @apply border-primary-400 ring-1 ring-primary-200;
  }

  :global(.editor-wrapper .cm-editor) {
    height: 84px;
    min-height: 48px;
    resize: vertical;
    overflow: hidden;
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
