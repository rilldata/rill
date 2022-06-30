<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import type { ColumnConfig } from "$lib/components/table-editor/ColumnConfig";

  export let value;
  export let index;
  export let column: ColumnConfig;
  export let isNull = false;

  const dispatch = createEventDispatcher();

  let editing = false;
</script>

<input
  id="model-title-input"
  on:input={() => (editing = true)}
  class="rounded pl-2 pr-2 cursor-pointer w-full"
  on:blur={() => {
    editing = false;
  }}
  on:change={(evt) => {
    dispatch("change", {
      value: evt.target.value,
      name: column.name,
      index,
    });
  }}
  value={value ?? ""}
/>
