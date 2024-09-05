<script lang="ts">
  import { EyeIcon, EyeOffIcon } from "lucide-svelte";
  import { onMount } from "svelte";
  import { slide } from "svelte/transition";
  import InfoCircle from "../icons/InfoCircle.svelte";
  import Tooltip from "../tooltip/Tooltip.svelte";
  import TooltipContent from "../tooltip/TooltipContent.svelte";
  import Select from "./Select.svelte";

  type InputEvent = {
    currentTarget: EventTarget & HTMLInputElement;
  };

  const voidFunction = () => {};

  export let value: string | undefined | null;
  export let id = "";
  export let label = "";
  export let description = "";
  export let errors: string | string[] | null | undefined = null;
  export let placeholder = "";
  export let hint = "";
  export let claimFocusOnMount = false;
  export let secret = false;
  export let autocomplete = false;
  export let alwaysShowError = false;
  export let optional = false;
  export let modes: {
    label: string;
    type: "select" | "text" | "combo";
    options: string[];
  }[] = [];
  export let onInput: (e: Event & InputEvent) => void = voidFunction;
  export let onChange: (e: Event & InputEvent) => void = voidFunction;
  export let inputType = "text";
  export let selected: string = "Simple";

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

  {#if modes.length}
    <div class="rounded-sm option-wrapper flex h-6 text-sm w-fit mb-1">
      {#each modes as { label, type } (label)}
        <button
          on:click={() => {
            inputType = type;
            selected = label;
          }}
          class="-ml-[1px] first-of-type:-ml-0 px-2 border first-of-type:rounded-l-sm last-of-type:rounded-r-sm"
          class:selected={selected === label}>{label}</button
        >
      {/each}
    </div>
  {/if}

  {#if inputType === "text"}
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
          class="toggle"
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
  {:else if inputType === "select"}
    <Select
      {value}
      options={modes
        .find(({ label }) => label === selected)
        ?.options.map((value) => ({ value }))}
      on:change={(e) => {
        value = e.detail;
      }}
    />
  {/if}

  {#if errors && (alwaysShowError || (!focus && value))}
    {#if typeof errors === "string"}
      <div in:slide={{ duration: 200 }} class="error">
        {errors}
      </div>
    {:else}
      {#each errors as error (error)}
        <div in:slide={{ duration: 200 }} class="error">
          {error}
        </div>
      {/each}
    {/if}
  {/if}

  {#if description}
    <div>{description}</div>
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
    @apply flex justify-center items-center;
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

  .toggle {
    @apply h-full aspect-square flex items-center justify-center;
  }

  .toggle:hover {
    @apply bg-primary-50 cursor-pointer;
  }

  .toggle:active {
    @apply bg-primary-100;
  }

  .option-wrapper > .selected {
    @apply border-primary-500 z-50 text-primary-500;
  }
</style>
