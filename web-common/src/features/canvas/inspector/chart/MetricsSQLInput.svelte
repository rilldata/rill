<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
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
    if (value.length === 1) return; // Don't remove last
    const updated = value.slice();
    updated.splice(idx, 1);
    onChange(updated);
  }
</script>

<div class="flex flex-col gap-y-4">
  {#each value as sql, idx (idx)}
    <div class="flex flex-col gap-y-2 relative group">
      <InputLabel
        hint={description}
        small
        label={(label ?? key) + (value.length > 1 ? ` (Query ${idx + 1})` : "")}
        id={key + "-" + idx}
      />
      <textarea
        class="w-full p-2 border border-gray-300 rounded-sm font-mono"
        rows="3"
        bind:value={value[idx]}
        on:blur={(e) => updateSQL(idx, e.target.value)}
        placeholder="SELECT * FROM metrics"
      />
      {#if value.length > 1}
        <Button
          type="plain"
          class="absolute top-0 right-0 mt-1 mr-1 opacity-70 group-hover:opacity-100 z-10"
          small
          on:click={() => removeQuery(idx)}
          label="Remove this query"
        >
          ✕
        </Button>
      {/if}
    </div>
  {/each}
  <Button type="dashed" wide on:click={addQuery} class="mt-2">
    + Add another query
  </Button>
</div>
