<script lang="ts">
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";

  export let key: string;
  export let label: string | undefined;
  export let description: string | undefined;
  export let value: string[] = [];
  export let onChange: (updatedSQLs: string[]) => void;

  // Ensure at least one editor is always shown
  if (!value || value.length === 0) value = [""];

  function updateSQL(idx: number, newSQL: string) {
    const updated = value.slice();
    updated[idx] = newSQL;
    onChange(updated);
  }

  function addQuery() {
    onChange([...value, ""]);
  }

  function removeQuery(idx: number) {
    if (value.length === 1) return;
    const updated = value.slice();
    updated.splice(idx, 1);
    onChange(updated);
  }
</script>

<div class="sql-input-container">
  {#if label}
    <InputLabel hint={description} small {label} id={key} />
  {/if}

  <div class="queries">
    {#each value as sql, idx (idx)}
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
        <textarea
          class="sql-textarea"
          bind:value={value[idx]}
          on:blur={() => updateSQL(idx, value[idx])}
          placeholder="SELECT * FROM metrics"
        />
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

  .sql-textarea {
    @apply w-full px-2.5 py-2 text-xs;
    @apply border border-gray-200 rounded-md;
    @apply resize-none outline-none;
    @apply transition-colors duration-150;
    font-family: "Source Code Variable", monospace;
    min-height: 48px;
    max-height: 120px;
    field-sizing: content;
  }

  .sql-textarea:focus {
    @apply border-primary-400 ring-1 ring-primary-200;
  }

  .sql-textarea::placeholder {
    @apply text-gray-300;
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
