<script lang="ts">
  import { PlusIcon, XIcon } from "lucide-svelte";
  import { tick } from "svelte";
  import InputLabel from "./InputLabel.svelte";

  export let id: string;
  export let label = "";
  export let hint: string | undefined = undefined;
  export let optional = false;
  export let value: Array<{ key: string; value: string }> = [];
  export let keyPlaceholder = "Header name";
  export let valuePlaceholder = "Value";

  let keyInputs: HTMLInputElement[] = [];

  // Defensive: coerce to array if value is undefined or wrong type
  $: entries = Array.isArray(value) ? value : [];
  $: duplicateKeys = findDuplicateKeys(entries);

  function findDuplicateKeys(
    entries: Array<{ key: string; value: string }>,
  ): Set<string> {
    const seen = new Set<string>();
    const dupes = new Set<string>();
    for (const entry of entries) {
      const k = entry.key.trim();
      if (!k) continue;
      if (seen.has(k)) dupes.add(k);
      seen.add(k);
    }
    return dupes;
  }

  async function addRow() {
    value = [...value, { key: "", value: "" }];
    await tick();
    keyInputs[value.length - 1]?.focus();
  }

  function removeRow(index: number) {
    value = value.filter((_, i) => i !== index);
  }

  function updateKey(index: number, newKey: string) {
    value[index] = { ...value[index], key: newKey };
    value = value;
  }

  function updateValue(index: number, newValue: string) {
    value[index] = { ...value[index], value: newValue };
    value = value;
  }
</script>

<div class="flex flex-col gap-y-1">
  {#if label}
    <InputLabel {id} {label} {hint} {optional} />
  {/if}

  {#each entries as entry, i (i)}
    <div class="flex items-center gap-1.5">
      <div class="kv-input-wrapper flex-1">
        <input
          bind:this={keyInputs[i]}
          type="text"
          placeholder={keyPlaceholder}
          value={entry.key}
          on:input={(e) => updateKey(i, e.currentTarget.value)}
          aria-label="Header name {i + 1}"
          class="kv-input"
        />
      </div>
      <div class="kv-input-wrapper flex-1">
        <input
          type="text"
          placeholder={valuePlaceholder}
          value={entry.value}
          on:input={(e) => updateValue(i, e.currentTarget.value)}
          aria-label="Header value {i + 1}"
          class="kv-input"
        />
      </div>
      <button
        type="button"
        class="remove-button"
        on:click={() => removeRow(i)}
        aria-label="Remove header {i + 1}"
      >
        <XIcon size="14px" />
      </button>
    </div>
    {#if duplicateKeys.has(entry.key.trim())}
      <div class="text-xs text-amber-600">
        Duplicate key "{entry.key.trim()}" — last value wins
      </div>
    {/if}
  {/each}

  <button type="button" class="add-button" on:click={addRow}>
    <PlusIcon size="14px" />
    Add header
  </button>
</div>

<style lang="postcss">
  .kv-input-wrapper {
    @apply border rounded-[2px] bg-input px-2;
    @apply flex items-center;
    height: 30px;
  }

  .kv-input-wrapper:focus-within {
    @apply border-primary-500 ring-2 ring-primary-100;
  }

  .kv-input {
    @apply bg-transparent outline-none border-0;
    @apply text-xs placeholder-fg-muted;
    @apply w-full h-full;
  }

  .remove-button {
    @apply text-fg-muted;
    @apply flex-none flex items-center justify-center;
    @apply cursor-pointer;
    width: 24px;
    height: 30px;
  }

  .remove-button:hover {
    @apply text-fg-primary;
  }

  .add-button {
    @apply flex items-center gap-1;
    @apply text-xs text-primary-500 font-medium;
    @apply cursor-pointer w-fit mt-0.5;
  }

  .add-button:hover {
    @apply text-primary-600;
  }
</style>
