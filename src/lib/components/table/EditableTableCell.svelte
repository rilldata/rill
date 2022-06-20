<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { ColumnConfig } from "$lib/components/table/pinnableUtils";

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
  class="bg-gray-100 border border-transparent border-2 hover:border-gray-400 rounded pl-2 pr-2 cursor-pointer"
  class:font-bold={editing === false}
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
