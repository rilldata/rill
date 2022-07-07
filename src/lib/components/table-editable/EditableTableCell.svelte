<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import type { ColumnConfig } from "$lib/components/table-editable/ColumnConfig";

  export let value;
  export let index;
  export let column: ColumnConfig;

  const dispatch = createEventDispatcher();

  let editing = false;
  const onchangeHandler = (evt) => {
    dispatch("change", {
      value: evt.target.value,
      name: column.name,
      index,
    });
  };
</script>

<input
  id="model-title-input"
  on:input={() => (editing = true)}
  class="rounded pl-2 pr-2 cursor-pointer w-full"
  class:font-bold={editing === false}
  on:blur={() => {
    editing = false;
  }}
  on:change={onchangeHandler}
  value={value ?? ""}
/>
