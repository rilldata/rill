<script lang="ts">
  import { IconButton } from "@rilldata/web-common/components/button";
  import { EyeIcon, EyeOffIcon } from "lucide-svelte";
  import { onMount } from "svelte";
  import { slide } from "svelte/transition";
  import FieldSwitcher from "./FieldSwitcher.svelte";
  import InputLabel from "./InputLabel.svelte";
  import Select from "./Select.svelte";

  const voidFunction = () => {};

  export let value: number | string | undefined;
  export let inputType: "text" | "number" = "text";
  export let id = "";
  export let label = "";
  export let title = ""; // Fallback title attribute if label is empty
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
  export let size: "sm" | "md" | "lg" | "xl" = "lg";
  export let labelGap = 1;
  export let selected: number = -1;
  export let full = false;
  export let multiline = false;
  export let fontFamily = "inherit";
  export let sameWidth = false;
  export let textClass = "text-xs";
  export let enableSearch = false;
  export let fields: string[] | undefined = [];
  export let disabled = false;
  export let link: string = "";
  export let lockable = false;
  export let capitalizeLabel = true;
  export let textInputPrefix = "";
  export let lockTooltip: string | undefined = undefined;
  export let disabledMessage = "No valid options";
  export let options:
    | { value: string; label: string; type?: string }[]
    | undefined = undefined;
  export let additionalClass = "";
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
  export let onEnter: (
    e: KeyboardEvent & {
      currentTarget: EventTarget & HTMLDivElement;
    },
  ) => void = voidFunction;
  export let onEscape: () => void = voidFunction;
  export let onFieldSwitch: (i: number, value: string) => void = voidFunction;

  let hitEnter = false;
  let showPassword = false;
  let inputElement: HTMLElement | undefined;
  let selectElement: HTMLButtonElement | undefined;
  let focus = false;

  $: type = secret && !showPassword ? "password" : inputType;

  $: hasValue = inputType === "number" ? !!value || value === 0 : !!value;

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
    if (hitEnter) {
      hitEnter = false;
      return;
    }
    focus = false;
    onBlur(e);
  }

  function onKeydown(
    e: KeyboardEvent & {
      currentTarget: EventTarget & HTMLDivElement;
    },
  ) {
    if (e.key === "Enter") {
      if (e.shiftKey) return;
      hitEnter = true;
      inputElement?.blur();
      onEnter(e);
    } else if (e.key === "Escape") {
      e.preventDefault();
      onEscape();
    }
  }
</script>

<div
  class="component-wrapper gap-y-{labelGap} {additionalClass}"
  class:w-full={full}
  style:width
>
  {#if label}
    <InputLabel
      {label}
      {optional}
      {id}
      {hint}
      {link}
      small={size === "sm"}
      capitalize={capitalizeLabel}
    >
      <slot name="mode-switch" slot="mode-switch" />
    </InputLabel>
  {/if}

  {#if fields && fields?.length > 1}
    <FieldSwitcher {fields} {selected} onClick={onFieldSwitch} />
  {/if}

  {#if !options}
    <div
      class="input-wrapper {textClass}"
      style:width
      class:error-input-wrapper={!!errors?.length}
      style:font-family={fontFamily}
      class:pl-2={!textInputPrefix}
    >
      {#if textInputPrefix}
        <div
          class:text-sm={size !== "xl"}
          class="{size} bg-neutral-100 items-center flex flex-none cursor-default line-clamp-1 text-gray-500 border-r border-gray-300 text-base px-2 mr-2"
        >
          {textInputPrefix}
        </div>
      {/if}

      {#if $$slots.icon}
        <span class="mr-1 flex-none">
          <slot name="icon" />
        </span>
      {/if}

      {#if multiline && typeof value !== "number"}
        <div
          {id}
          contenteditable
          class="multiline-input"
          class:pointer-events-none={disabled}
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
          title={label || title}
          {id}
          {type}
          {placeholder}
          name={id}
          class={size}
          {disabled}
          value={value ?? (inputType === "number" ? null : "")}
          autocomplete={autocomplete ? "on" : "off"}
          bind:this={inputElement}
          on:input={(e) => {
            if (inputType === "number") {
              if (e.currentTarget.value === "") {
                value = undefined;
              } else {
                value = e.currentTarget.valueAsNumber;
              }
              return;
            }
            value = e.currentTarget.value;
            onInput(value, e);
          }}
          on:keydown={onKeydown}
          on:blur={onElementBlur}
          on:focus={() => (focus = true)}
        />
      {/if}
      {#if secret}
        <IconButton
          size={20}
          disableHover
          ariaLabel={showPassword ? "Hide password" : "Show password"}
          on:click={() => {
            showPassword = !showPassword;
          }}
        >
          {#if showPassword}
            <EyeOffIcon size="14px" class="text-muted-foreground" />
          {:else}
            <EyeIcon size="14px" class="text-muted-foreground" />
          {/if}
        </IconButton>
      {/if}
    </div>
  {:else if typeof value !== "number"}
    <Select
      {disabled}
      {enableSearch}
      ringFocus
      {sameWidth}
      {id}
      {lockable}
      {lockTooltip}
      bind:selectElement
      bind:value
      {options}
      {onChange}
      {size}
      fontSize={size === "sm" ? 12 : 14}
      {truncate}
      placeholder={disabled ? disabledMessage : placeholder}
    />
  {/if}

  {#if errors && (alwaysShowError || (!focus && hasValue))}
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

  {#if $$slots.description || description}
    <div class="description">
      {#if $$slots.description}
        <slot name="description" />
      {:else}
        {description}
      {/if}
    </div>
  {/if}
</div>

<style lang="postcss">
  .component-wrapper {
    @apply flex flex-col h-fit justify-center;
  }

  .sm {
    height: 24px;
    font-size: 12px;
  }

  .md {
    height: 26px;
  }

  .lg {
    height: 30px;
  }

  .xl {
    height: 38px;
    font-size: 16px;
  }

  .input-wrapper {
    @apply overflow-hidden;
    @apply flex justify-center items-center pr-1;
    @apply bg-surface justify-center;
    @apply border border-gray-300 rounded-[2px];
    @apply cursor-pointer;
    @apply h-fit w-fit;
  }

  .input-wrapper:has(input:disabled) {
    @apply bg-gray-50 border-gray-200 cursor-not-allowed;
  }

  input,
  .multiline-input {
    @apply bg-surface p-0;
    @apply size-full;
    @apply outline-none border-0;
    @apply cursor-text;
    vertical-align: middle;
  }

  input:disabled {
    @apply bg-gray-50 text-gray-500 cursor-not-allowed;
  }

  input {
    @apply truncate;
  }

  .multiline-input {
    @apply h-fit w-full max-h-32 overflow-y-auto;
    @apply py-1;
    line-height: 1.58;
    word-wrap: break-word;
  }

  .input-wrapper:focus-within {
    @apply border-primary-500;
    @apply ring-2 ring-primary-100;
  }

  .error-input-wrapper.input-wrapper {
    @apply border-red-600;
    @apply ring-1 ring-transparent;
  }

  .error {
    @apply text-red-500 text-xs;
  }

  .toggle {
    @apply h-full aspect-square flex items-center justify-center;
    @apply px-1.5;
    @apply -mr-0.5;
  }

  .toggle:hover {
    @apply bg-primary-50 cursor-pointer;
  }

  .toggle:active {
    @apply bg-primary-100;
  }

  .description {
    @apply text-xs text-gray-500;
  }
</style>
