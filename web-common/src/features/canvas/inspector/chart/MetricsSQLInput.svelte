<script lang="ts">
  import { IconButton } from "@rilldata/web-common/components/button";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import MinusIcon from "@rilldata/web-common/components/icons/MinusIcon.svelte";

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

<div class="flex flex-col gap-y-2">
  {#each value as sql, idx (idx)}
    <div class="flex flex-col gap-y-2 relative group">
      <div class="flex items-center justify-between">
        <InputLabel
          hint={description}
          small
          label={(label ?? key) +
            (value.length > 1 ? ` (Query ${idx + 1})` : "")}
          id={key + "-" + idx}
        />
        {#if value.length > 1}
          <IconButton
            rounded
            on:click={() => removeQuery(idx)}
            ariaLabel="Remove this query"
          >
            <MinusIcon />
          </IconButton>
        {/if}
      </div>
      <textarea
        class="w-full p-2 border border-gray-300 rounded-sm source-code"
        rows="2"
        bind:value={value[idx]}
        on:blur={(e) => updateSQL(idx, value[idx])}
        placeholder="SELECT * FROM metrics"
      />
    </div>
  {/each}
  <Button type="dashed" small on:click={addQuery} class="mt-2">
    + Add another query
  </Button>
</div>

<style>
  .source-code {
    font-family: "Source Code Variable", monospace;
  }
</style>
