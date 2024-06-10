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
  export let claimFocusOnMount = false;
  export let isSecret = false;

  let inputElement;

  if (claimFocusOnMount) {
    onMount(() => {
      inputElement.focus();
    });
  }
</script>

<div class="label-wrapper">
  <label for={id}>{label}</label>
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
</div>
<!-- This conditional for `isSecret` is necessary because Svelte does not allow a dynamic `type` for an `input` using two-way binding. -->
{#if isSecret}
  <input
    type="password"
    bind:this={inputElement}
    bind:value
    on:input
    on:change
    {id}
    name={id}
    {placeholder}
    autocomplete="off"
  />
{:else}
  <input
    type="text"
    bind:this={inputElement}
    bind:value
    on:input
    on:change
    {id}
    name={id}
    {placeholder}
    autocomplete="off"
  />
{/if}
{#if error}
  <div in:slide={{ duration: 200 }} class="error">
    {error}
  </div>
{/if}

<style lang="postcss">
  .label-wrapper {
    @apply flex items-center gap-x-1;
    @apply pb-2;
  }

  label {
    @apply text-sm font-medium text-gray-800;
  }

  input {
    @apply px-3 py-1 w-full;
    @apply border border-gray-300 rounded-sm;
    @apply text-xs;
    @apply cursor-pointer;
  }

  input:focus {
    @apply outline-primary-500;
  }

  .error {
    @apply pl-1 pt-1;
    @apply text-red-500 text-xs;
  }
</style>
