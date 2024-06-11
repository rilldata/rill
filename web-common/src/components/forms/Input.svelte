<script lang="ts">
  import { onMount } from "svelte";
  import { slide } from "svelte/transition";
  import InfoCircle from "../icons/InfoCircle.svelte";
  import Tooltip from "../tooltip/Tooltip.svelte";
  import TooltipContent from "../tooltip/TooltipContent.svelte";
  import { EyeOffIcon, EyeIcon } from "lucide-svelte";

  type InputEvent = {
    currentTarget: EventTarget & HTMLInputElement;
  };

  const voidFunction = () => {};

  export let value: string;
  export let id = "";
  export let label = "";
  export let error: string | null = null;
  export let placeholder = "";
  export let hint = "";
  export let claimFocusOnMount = false;
  export let secret = false;
  export let autocomplete = false;
  export let alwaysShowError = false;
  export let optional = false;
  export let onInput: (e: Event & InputEvent) => void = voidFunction;
  export let onChange: (e: Event & InputEvent) => void = voidFunction;

  let showPassword = false;
  let inputElement: HTMLInputElement;
  let focus = false;

  $: type = secret && !showPassword ? "password" : "text";

  onMount(() => {
    if (claimFocusOnMount) {
      inputElement.focus();
    }
  });
</script>

<div class="flex flex-col gap-y-1">
  {#if label}
    <div class="label-wrapper">
      <label for={id}>
        {label}
        {#if optional}
          <span class="text-gray-500 text-sm">(optional)</span>
        {/if}
      </label>
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
  {/if}

  <div class="input-wrapper">
    <input
      {id}
      {type}
      {placeholder}
      name={id}
      value={value ?? ""}
      autocomplete={autocomplete ? "on" : "off"}
      bind:this={inputElement}
      on:change={onChange}
      on:input={(e) => {
        value = e.currentTarget.value;
        onInput(e);
      }}
      on:blur={() => (focus = false)}
      on:focus={() => (focus = true)}
    />
    {#if secret}
      <button
        type="button"
        aria-label={showPassword ? "Hide password" : "Show password"}
        on:click={() => {
          showPassword = !showPassword;
        }}
      >
        {#if showPassword}
          <EyeOffIcon size="14px" class="stroke-primary-600" />
        {:else}
          <EyeIcon size="14px" class="stroke-primary-600" />
        {/if}
      </button>
    {/if}
  </div>

  {#if error && (alwaysShowError || (!focus && value))}
    <div in:slide={{ duration: 200 }} class="error">
      {error}
    </div>
  {/if}
</div>

<style lang="postcss">
  .label-wrapper {
    @apply flex items-center gap-x-1;
  }

  label {
    @apply text-sm font-medium text-gray-800;
  }

  input {
    @apply size-full outline-none border-0;
  }

  .input-wrapper {
    @apply flex justify-center items-center overflow-hidden;
    @apply h-8 pl-2 w-full;

    @apply border border-gray-300 rounded-sm;
    @apply text-xs;
    @apply cursor-pointer;
  }

  .input-wrapper:focus-within {
    @apply border-primary-500;
  }

  .error {
    @apply text-red-500 text-xs;
  }

  button {
    @apply h-full aspect-square  flex items-center justify-center;
  }

  button:hover {
    @apply bg-primary-50 cursor-pointer;
  }

  button:active {
    @apply bg-primary-100;
  }
</style>
