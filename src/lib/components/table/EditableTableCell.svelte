<script lang="ts">
  import Tooltip from "../tooltip/Tooltip.svelte";
  import TooltipContent from "../tooltip/TooltipContent.svelte";
  import EditIcon from "$lib/components/icons/EditIcon.svelte";
  import { createEventDispatcher } from "svelte";

  export let value;
  export let type;
  export let name;
  export let index = undefined;
  export let isNull = false;

  const dispatch = createEventDispatcher();

  let editing = false;
  let activeCell = false;
</script>

<Tooltip location="top" distance={16}>
  <td
    on:mouseover={() => {
      dispatch("inspect", index);
      activeCell = true;
    }}
    on:mouseout={() => {
      activeCell = false;
    }}
    on:focus={() => {
      dispatch("inspect", index);
      activeCell = true;
    }}
    on:blur={() => {
      activeCell = false;
    }}
    title={value}
    class="
        p-2
        pl-4
        pr-4
        border
        border-gray-200
        {activeCell && 'bg-gray-200'}
    "
    style:width="var(--table-column-width-{name}, 210px)"
    style:max-width="var(--table-column-width-{name}, 210px)"
  >
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
          name,
          index,
        });
      }}
      value={value ?? ""}
    />
  </td>
  <TooltipContent slot="tooltip-content">
    <div class="flex items-center"><EditIcon size=".75em" />Edit {name}</div>
  </TooltipContent>
</Tooltip>
