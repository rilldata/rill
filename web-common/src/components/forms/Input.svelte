<script lang="ts">
  import { EyeIcon, EyeOffIcon } from "lucide-svelte";
  import { onMount } from "svelte";
  import { slide } from "svelte/transition";
  import InfoCircle from "../icons/InfoCircle.svelte";
  import Tooltip from "../tooltip/Tooltip.svelte";
  import TooltipContent from "../tooltip/TooltipContent.svelte";
  import Select from "./Select.svelte";

  const voidFunction = () => {};

  export let value: string | undefined;
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
  export let fields: string[] | undefined = [];
  export let options: string[] | undefined = [];
  export let onInput: (
    newValue: string,
    e: Event & {
      currentTarget: EventTarget & HTMLElement;
    },
  ) => void = voidFunction;
  export let onChange: (newValue: string) => void = voidFunction;
  export let onBlur: (
    e: FocusEvent & {
      currentTarget: EventTarget & HTMLDivElement;
    },
  ) => void = voidFunction;
  export let selected: number = -1;
  export let full = false;
  export let multiline = false;
  export let fontFamily = "inherit";
  export let sameWidth = false;
  export let ringFocus = false;

  let showPassword = false;
  let inputElement: HTMLElement | undefined;
  let selectElement: HTMLButtonElement | undefined;
  let focus = false;

  $: type = secret && !showPassword ? "password" : "text";

  onMount(() => {
    if (claimFocusOnMount) {
      if (inputElement) {
        inputElement.focus();
      } else if (selectElement) {
        selectElement.focus();
      }
    }
  });

  function onElementBlur(
    e: FocusEvent & { currentTarget: EventTarget & HTMLDivElement },
  ) {
    console.log("blur");
    focus = false;
    onBlur(e);
  }
</script>

<div class="flex flex-col gap-y-1" class:w-full={full}>
  {#if label}
    <div class="label-wrapper">
      <label for={id} class="line-clamp-1">
        {label}
        {#if optional}
          <span class="text-gray-500 text-[12px] font-normal">(optional)</span>
        {/if}
      </label>
      {#if hint}
        <Tooltip location="right" alignment="middle" distance={8}>
          <div class="text-gray-500">
            <InfoCircle size="13px" />
          </div>
          <TooltipContent maxWidth="400px" slot="tooltip-content">
            {@html hint}
          </TooltipContent>
        </Tooltip>
      {/if}
    </div>
  {/if}

  {#if fields && fields?.length > 1}
    <div class="rounded-sm option-wrapper flex h-6 text-sm w-fit mb-1">
      {#each fields as field, i (field)}
        <button
          on:click={() => {
            selected = i;
          }}
          class="-ml-[1px] first-of-type:-ml-0 px-2 border border-gray-300 first-of-type:rounded-l-[2px] last-of-type:rounded-r-[2px]"
          class:selected={selected === i}
        >
          {field}
        </button>
      {/each}
    </div>
  {/if}

  {#if !options?.length}
    <div
      class="input-wrapper overflow-hidden"
      class:ring-focus={ringFocus}
      style:font-family={fontFamily}
    >
      {#if multiline}
        <div
          {id}
          contentEditable="true"
          class="multiline-input"
          {placeholder}
          role="textbox"
          aria-multiline="true"
          bind:this={inputElement}
          on:input={(e) => {
            value = e.currentTarget.textContent ?? "";
            onInput(value, e);
          }}
          on:blur={onElementBlur}
          on:focus={() => (focus = true)}
        >
          {value ?? ""}
        </div>
      {:else}
        <input
          {id}
          {type}
          {placeholder}
          name={id}
          value={value ?? ""}
          autocomplete={autocomplete ? "on" : "off"}
          bind:this={inputElement}
          on:input={(e) => {
            value = e.currentTarget.value;
            onInput(value, e);
          }}
          on:blur={onElementBlur}
          on:focus={() => (focus = true)}
        />
      {/if}
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
  {:else if options.length}
    <Select
      {ringFocus}
      {sameWidth}
      {id}
      bind:selectElement
      bind:value
      options={options.map((value) => ({ value, label: value })) ?? []}
      {onChange}
      fontSize={14}
      {placeholder}
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

  input,
  .multiline-input {
    @apply size-full;
    @apply outline-none border-0;
  }

  .multiline-input {
    @apply overflow-auto break-words py-[5px];
    @apply text-sm;
  }

  input {
    @apply h-[30px] py-0;
  }

  .input-wrapper {
    @apply flex justify-center items-center pl-2;
    @apply w-full;
    @apply border border-gray-300 rounded-[2px];
    @apply text-xs;
    @apply cursor-pointer;
    @apply min-h-8 h-fit;
  }

  .input-wrapper:focus-within {
    @apply border-primary-500;
  }

  .ring-focus:focus-within {
    @apply ring-2 ring-primary-100;
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
