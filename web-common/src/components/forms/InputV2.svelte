<script lang="ts">
  import { onMount } from "svelte";
  import { slide } from "svelte/transition";
  import InfoCircle from "../icons/InfoCircle.svelte";
  import Tooltip from "../tooltip/Tooltip.svelte";
  import TooltipContent from "../tooltip/TooltipContent.svelte";

  export let id = "";
  export let label = "";
  export let error: string;
  export let value: string;
  export let placeholder = "";
  export let hint = "";
  export let optional = false;
  export let claimFocusOnMount = false;

  let inputElement;

  if (claimFocusOnMount) {
    onMount(() => {
      inputElement.focus();
    });
  }

  function handleKeyDown(event: KeyboardEvent) {
    if (event.key === "Enter") {
      event.preventDefault();
    }
  }
</script>

<div class="flex flex-col gap-y-2">
  {#if label}
    <div class="flex items-center gap-x-1">
      {#if label}
        <label for={id} class="text-gray-800 text-sm font-medium">
          {label}
        </label>
      {/if}
      {#if hint}
        <Tooltip location="right" alignment="middle" distance={8}>
          <div class="text-gray-500" style="transform:translateY(-.5px)">
            <InfoCircle size="13px" />
          </div>
          <TooltipContent maxWidth="400px" slot="tooltip-content">
            {@html hint}
          </TooltipContent>
        </Tooltip>
      {/if}
      {#if optional}
        <span class="text-gray-500 text-sm">(optional)</span>
      {/if}
    </div>
  {/if}
  <input
    bind:this={inputElement}
    bind:value
    on:input
    on:change
    on:keydown={handleKeyDown}
    {id}
    name={id}
    type="text"
    {placeholder}
    autocomplete="off"
    class="bg-white rounded-sm border border-gray-300 px-3 py-[5px] h-8 cursor-pointer focus:outline-blue-500 w-full text-xs {error &&
      'border-red-500'}"
  />
  {#if error}
    <div in:slide={{ duration: 200 }} class="text-red-500 text-sm py-px">
      {error}
    </div>
  {/if}
</div>
