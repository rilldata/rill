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
  export let truncate = false;
  export let width: string = "100%";
  export let size: "sm" | "md" | "lg" = "lg";
  export let selected: number = -1;
  export let full = false;
  export let multiline = false;
  export let fontFamily = "inherit";
  export let sameWidth = false;
  export let textClass = "text-xs";
  export let enableSearch = false;
  export let fields: string[] | undefined = [];
  export let noOptionsMessage = "No valid options";
  export let options:
    | { value: string; label: string; type?: string }[]
    | undefined = undefined;
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
  export let onEnter: () => void = voidFunction;
  export let onEscape: () => void = voidFunction;

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
    focus = false;
    onBlur(e);
  }

  function onKeydown(
    e: KeyboardEvent & {
      currentTarget: EventTarget & HTMLDivElement;
    },
  ) {
    if (e.key === "Enter") {
      e.preventDefault();
      onEnter();
    } else if (e.key === "Escape") {
      e.preventDefault();
      onEscape();
    }
  }
</script>

<div class="component-wrapper" class:w-full={full} style:width>
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
    <div class="option-wrapper">
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

  {#if !options}
    <div
      class="input-wrapper {textClass}"
      style:width
      style:font-family={fontFamily}
    >
      {#if $$slots.icon}
        <span class="mr-1 flex-none">
          <slot name="icon" />
        </span>
      {/if}

      {#if multiline}
        <div
          {id}
          contenteditable
          class="multiline-input"
          {placeholder}
          role="textbox"
          tabindex="0"
          aria-multiline="true"
          bind:this={inputElement}
          bind:textContent={value}
          on:keydown={onKeydown}
          on:blur={onElementBlur}
          on:focus={() => (focus = true)}
        />
      {:else}
        <input
          {id}
          {type}
          {placeholder}
          name={id}
          class={size}
          value={value ?? ""}
          autocomplete={autocomplete ? "on" : "off"}
          bind:this={inputElement}
          on:input={(e) => {
            value = e.currentTarget.value;
            onInput(value, e);
          }}
          on:keydown={onKeydown}
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
      {enableSearch}
      ringFocus
      {sameWidth}
      {id}
      bind:selectElement
      bind:value
      options={options ?? []}
      {onChange}
      fontSize={14}
      {truncate}
      {placeholder}
    />
  {:else}
    <div class="text-sm h-8 items-center flex text-gray-500 italic">
      {noOptionsMessage}
    </div>
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
  .component-wrapper {
    @apply flex  flex-col gap-y-1 h-fit justify-center;
  }

  .label-wrapper {
    @apply flex items-center gap-x-1;
  }

  label {
    @apply text-sm font-medium text-gray-800;
  }

  .option-wrapper {
    @apply flex h-6 text-sm w-fit mb-1 rounded-[2px];
  }

  .sm {
    height: 24px;
  }

  .md {
    height: 26px;
  }

  .lg {
    height: 30px;
  }

  .input-wrapper {
    @apply overflow-hidden;
    @apply flex justify-center items-center px-2;
    @apply bg-background justify-center;
    @apply border border-gray-300 rounded-[2px];
    @apply cursor-pointer;
    @apply h-fit w-fit;
  }

  input,
  .multiline-input {
    @apply p-0 bg-transparent;
    @apply size-full;
    @apply outline-none border-0;
    @apply cursor-text;
    vertical-align: middle;
  }

  .multiline-input {
    @apply py-1;
    line-height: 1.6;
  }

  .multiline-input {
    @apply overflow-auto break-words;
    @apply h-fit min-h-fit;
  }

  .input-wrapper:focus-within {
    @apply border-primary-500;
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
