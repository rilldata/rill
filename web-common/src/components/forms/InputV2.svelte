<script lang="ts">
  import { createEventDispatcher, onMount } from "svelte";
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

  const dispatch = createEventDispatcher();

  let inputElement: HTMLInputElement;
  let focus = false;

  if (claimFocusOnMount) {
    onMount(() => {
      inputElement.focus();
    });
  }

  function handleKeyDown(event: KeyboardEvent) {
    if (event.key === "Enter") {
      event.preventDefault();
      inputElement.blur();
      dispatch("enter-pressed");
    }
  }
</script>

<div class="flex flex-col gap-y-2">
  <div class="flex items-center gap-x-1">
    <label class="text-gray-800 text-sm font-medium" for={id}>{label}</label>
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
  <input
    {id}
    type="text"
    {placeholder}
    name={id}
    autocomplete="off"
    class:error={error && value}
    on:change
    on:input
    on:keydown={handleKeyDown}
    on:focus={() => (focus = true)}
    on:blur={() => (focus = false)}
    bind:this={inputElement}
    bind:value
  />
  {#if error && !focus && value}
    <div in:slide={{ duration: 200 }} class="text-red-500 text-sm py-px">
      {error}
    </div>
  {/if}
</div>

<style lang="postcss">
  input {
    @apply w-full h-8 rounded-sm;
    @apply px-3 py-[5px];
    @apply text-xs;
    @apply bg-white border border-gray-300;
  }

  input:focus {
    @apply outline-primary-500;
  }
  .error:not(:focus) {
    @apply border-red-500;
  }
</style>
