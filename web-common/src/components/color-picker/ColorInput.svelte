<script lang="ts">
  import {
    isValidColor,
    stringColorToHsl,
  } from "@rilldata/web-common/components/color-picker/util";
  import * as Popover from "@rilldata/web-common/components/popover";
  import InputLabel from "../forms/InputLabel.svelte";
  import WarningIcon from "../icons/WarningIcon.svelte";
  import ColorSlider from "./ColorSlider.svelte";

  export let stringColor: string | undefined;
  export let label: string;
  export let disabled = false;
  export let onChange: (color: string) => void;
  export let small = false;
  export let allowLightnessControl = false;
  export let labelFirst = false;

  let open = false;

  // Use a reactive block (not destructuring assignment) so that hue/saturation/lightness
  // are plain state variables, not derived. Derived values are read-only in Svelte 5 compat
  // and can't be updated via bind:value on the sliders.
  let hue = 0;
  let saturation = 100;
  let lightness = 50;
  $: {
    const parsed = stringColorToHsl(stringColor);
    hue = parsed.h;
    saturation = parsed.s;
    lightness = parsed.l;
  }

  $: hsl = `hsl(${hue}, ${saturation}%, ${lightness}%)`;

  $: isColorValid = isValidColor(stringColor);
</script>

<svelte:window
  onkeydown={(e) => {
    if (e.key === "Escape" || e.key === "Enter") {
      open = false;
    }
  }}
/>

<div
  class="color-wrapper bg-input"
  class:small
  class:pointer-events-none={disabled}
  class:opacity-50={disabled}
  class:text-fg-secondary={disabled}
>
  {#if labelFirst}
    <p class:small class:text-fg-secondary={disabled} class="label-first">
      {label}
    </p>
    <div class="input-button-container">
      <input
        class:small
        class:text-right={labelFirst}
        class:text-red-500={!isColorValid}
        bind:value={stringColor}
        {disabled}
        onkeydown={(e) => {
          if (e.key === "Enter") {
            e.currentTarget.blur();
          }
        }}
        onblur={() => {
          if (stringColor) {
            onChange(stringColor);
          }
        }}
      />

      <Popover.Root bind:open>
        <Popover.Trigger>
          {#snippet child({ props })}
            <button
              {...props}
              class="trigger"
              class:error-trigger={!isColorValid}
              class:open
              style:--hsl={hsl}
            >
              {#if !isColorValid}
                <WarningIcon size="0.875rem" color="#f59e0b" />
              {/if}
            </button>
          {/snippet}
        </Popover.Trigger>

        <Popover.Content
          class="w-[270px] space-y-1.5"
          align="start"
          sideOffset={10}
        >
          <div class="space-y-0.5 -mt-1">
            <InputLabel label="Hue" id="hue" />
            <ColorSlider
              mode="hue"
              bind:value={hue}
              {hue}
              color={hsl}
              onChange={() => {
                stringColor = `hsl(${hue}, ${saturation}%, ${lightness}%)`;
                onChange(stringColor);
              }}
            />
          </div>
          <div class="space-y-0.5">
            <InputLabel label="Saturation" id="saturation" />
            <ColorSlider
              mode="saturation"
              bind:value={saturation}
              {hue}
              color={hsl}
              onChange={() => {
                stringColor = `hsl(${hue}, ${saturation}%, ${lightness}%)`;
                onChange(stringColor);
              }}
            />
          </div>
          {#if allowLightnessControl}
            <div class="space-y-0.5">
              <InputLabel label="Lightness" id="lightness" />
              <ColorSlider
                mode="lightness"
                bind:value={lightness}
                {hue}
                color={hsl}
                onChange={() => {
                  stringColor = `hsl(${hue}, ${saturation}%, ${lightness}%)`;
                  onChange(stringColor);
                }}
              />
            </div>
          {/if}
        </Popover.Content>
      </Popover.Root>
    </div>
  {:else}
    <Popover.Root bind:open>
      <Popover.Trigger>
        {#snippet child({ props })}
          <button
            {...props}
            class="trigger"
            class:error-trigger={!isColorValid}
            class:open
            style:--hsl={hsl}
          >
            {#if !isColorValid}
              <WarningIcon size="0.875rem" color="#f59e0b" />
            {/if}
          </button>
        {/snippet}
      </Popover.Trigger>

      <Popover.Content
        class="w-[270px] space-y-1.5"
        align="start"
        sideOffset={10}
      >
        <div class="space-y-0.5 -mt-1">
          <InputLabel label="Hue" id="hue" />
          <ColorSlider
            mode="hue"
            bind:value={hue}
            {hue}
            color={hsl}
            onChange={() => {
              stringColor = `hsl(${hue}, ${saturation}%, ${lightness}%)`;
              onChange(stringColor);
            }}
          />
        </div>
        <div class="space-y-0.5">
          <InputLabel label="Saturation" id="saturation" />
          <ColorSlider
            mode="saturation"
            bind:value={saturation}
            {hue}
            color={hsl}
            onChange={() => {
              stringColor = `hsl(${hue}, ${saturation}%, ${lightness}%)`;
              onChange(stringColor);
            }}
          />
        </div>
        {#if allowLightnessControl}
          <div class="space-y-0.5">
            <InputLabel label="Lightness" id="lightness" />
            <ColorSlider
              mode="lightness"
              bind:value={lightness}
              {hue}
              color={hsl}
              onChange={() => {
                stringColor = `hsl(${hue}, ${saturation}%, ${lightness}%)`;
                onChange(stringColor);
              }}
            />
          </div>
        {/if}
      </Popover.Content>
    </Popover.Root>

    <input
      class:small
      class:text-red-500={!isColorValid}
      bind:value={stringColor}
      {disabled}
      onkeydown={(e) => {
        if (e.key === "Enter") {
          e.currentTarget.blur();
        }
      }}
      onblur={() => {
        if (stringColor) {
          onChange(stringColor);
        }
      }}
    />

    <p class:small class:text-fg-secondary={disabled}>{label}</p>
  {/if}
</div>

<style lang="postcss">
  .trigger {
    @apply h-full aspect-square;
    @apply rounded-[2px] overflow-hidden flex-none;
    background-color: var(--hsl);
  }

  .trigger:hover,
  .trigger.open {
    border: 1.5px solid;
    border-color: color-mix(in oklch, var(--hsl), black 40%);
  }

  .error-trigger {
    @apply bg-red-50;
    @apply flex items-center justify-center;
    background-color: #fef2f2 !important;
  }

  .error-trigger:hover,
  .error-trigger.open {
    @apply bg-red-100;
    border: 1.5px solid #fca5a5;
  }

  .color-wrapper {
    @apply py-[5px] px-3;
    @apply h-8 w-full border rounded-[2px];
    @apply flex gap-x-3;
  }

  .color-wrapper.small {
    @apply h-6 text-xs;
  }

  .color-wrapper:focus-within {
    @apply border-primary-500 ring-2 ring-primary-100;
  }

  p {
    @apply text-sm;
  }

  p.small {
    @apply text-xs;
  }

  input {
    @apply w-full text-sm;
    @apply outline-none border-0 bg-transparent;
  }

  input.small {
    @apply text-xs;
  }

  .label-first {
    max-width: 60%;
    @apply flex-shrink-0;
    @apply truncate;
    @apply m-0 leading-tight;
    @apply self-center;
  }

  .input-button-container {
    @apply flex gap-x-3 flex-1;
    @apply min-w-0;
  }

  .input-button-container input {
    @apply flex-1 min-w-0;
  }
</style>
