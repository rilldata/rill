<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { slide } from "svelte/transition";
  import { Button, IconButton } from "../button";
  import Add from "../icons/Add.svelte";
  import InfoCircle from "../icons/InfoCircle.svelte";
  import Trash from "../icons/Trash.svelte";
  import Tooltip from "../tooltip/Tooltip.svelte";
  import TooltipContent from "../tooltip/TooltipContent.svelte";

  export let id = "";
  export let label = "";
  export let values: any[];
  export let errors: any[];
  // The accessorKey is necessary due to the way svelte-forms-lib works with arrays.
  // See: https://svelte-forms-lib-sapper-docs.vercel.app/array
  export let accessorKey: string;
  export let placeholder = "";
  export let hint = "";
  export let addItemLabel = "Add item";

  const dispatch = createEventDispatcher();

  function handleKeyDown(event: KeyboardEvent) {
    if (event.key === "Enter") {
      event.preventDefault();
    }
  }
</script>

<div class="flex flex-col gap-y-2.5">
  {#if label}
    <div class="flex items-center gap-x-1">
      <label for={id} class="text-gray-800 text-sm font-medium">{label}</label>
      {#if hint}
        <Tooltip location="right" alignment="middle" distance={8}>
          <div class="text-gray-500" style="transform:translateY(-.5px)">
            <InfoCircle size="13px" />
          </div>
          <TooltipContent maxWidth="400px" slot="tooltip-content">
            {hint}
          </TooltipContent>
        </Tooltip>
      {/if}
    </div>
  {/if}
  <div
    class="flex flex-col gap-y-4 max-h-[200px] pl-1 pr-4 py-1 overflow-y-auto"
  >
    {#each values as _, i}
      <div class="flex flex-col gap-y-2">
        <div class="flex gap-x-2 items-center">
          <input
            bind:value={values[i][accessorKey]}
            id="{id}.{i}.{accessorKey}"
            autocomplete="off"
            {placeholder}
            class="bg-white rounded-sm border border-gray-300 px-3 py-[5px] h-8 cursor-pointer focus:outline-primary-500 w-full text-xs {errors[
              i
            ]?.accessorKey && 'border-red-500'}"
            on:keydown={handleKeyDown}
          />
          <IconButton
            on:click={() =>
              dispatch("remove-item", {
                index: i,
              })}
          >
            <Trash size="16px" className="text-gray-500 cursor-pointer" />
          </IconButton>
        </div>
        {#if errors[i]?.[accessorKey]}
          <div in:slide={{ duration: 200 }} class="text-red-500 text-sm py-px">
            {errors[i][accessorKey]}
          </div>
        {/if}
      </div>
    {/each}
    <Button on:click={() => dispatch("add-item")} type="dashed">
      <div class="flex gap-x-2">
        <Add className="text-gray-700" />
        {addItemLabel}
      </div>
    </Button>
  </div>
</div>
