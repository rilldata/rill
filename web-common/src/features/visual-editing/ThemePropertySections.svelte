<script lang="ts">
  import {
    isValidColor,
    stringColorToHsl,
  } from "@rilldata/web-common/components/color-picker/util";
  import ColorSlider from "@rilldata/web-common/components/color-picker/ColorSlider.svelte";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import { Chip } from "@rilldata/web-common/components/chip";
  import * as Popover from "@rilldata/web-common/components/popover";
  import { slide } from "svelte/transition";
  import { THEME_SECTIONS } from "./theme-property-config";

  export let values: Record<string, string>;
  export let onPropertyChange: (key: string, value: string) => void;

  // Track open/closed state per section
  let openSections: Record<string, boolean> = {};
  for (const section of THEME_SECTIONS) {
    openSections[section.id] = section.defaultOpen ?? false;
  }

  // Color picker state per property (keyed by prop key)
  let colorState: Record<string, { h: number; s: number; l: number }> = {};

  function getHsl(key: string, stringColor: string | undefined) {
    if (!colorState[key]) {
      const parsed = stringColorToHsl(stringColor);
      colorState[key] = { h: parsed.h, s: parsed.s, l: parsed.l };
    }
    return colorState[key];
  }

  function updateColorState(key: string, stringColor: string | undefined) {
    const parsed = stringColorToHsl(stringColor);
    colorState[key] = { h: parsed.h, s: parsed.s, l: parsed.l };
  }

  function hslString(state: { h: number; s: number; l: number }) {
    return `hsl(${state.h}, ${state.s}%, ${state.l}%)`;
  }
</script>

<div class="sections">
  {#each THEME_SECTIONS as section (section.id)}
    <div class="section">
      <Chip
        gray
        caret
        active={openSections[section.id]}
        label={section.label}
        slideDuration={0}
        onclick={() => {
          openSections[section.id] = !openSections[section.id];
        }}
      >
        <span slot="body" class="text-xs font-semibold">{section.label}</span>
      </Chip>

      {#if openSections[section.id]}
        <div class="properties" transition:slide={{ duration: 150 }}>
          {#each section.properties as prop (prop.key)}
            {@const colorValue = values[prop.key] ?? ""}
            {@const cs = getHsl(prop.key, colorValue)}
            <div class="property-row">
              <Popover.Root
                onOpenChange={(open) => {
                  if (open) updateColorState(prop.key, colorValue);
                }}
              >
                <Popover.Trigger>
                  {#snippet child({ props: triggerProps })}
                    <button
                      {...triggerProps}
                      class="swatch"
                      class:swatch-empty={!colorValue}
                      style:background-color={colorValue || "#e5e7eb"}
                    ></button>
                  {/snippet}
                </Popover.Trigger>
                <Popover.Content
                  class="w-[270px] space-y-1.5"
                  align="start"
                  sideOffset={10}
                >
                  <div class="space-y-0.5 -mt-1">
                    <InputLabel label="Hue" id="hue-{prop.key}" />
                    <ColorSlider
                      mode="hue"
                      bind:value={cs.h}
                      hue={cs.h}
                      color={hslString(cs)}
                      onChange={() => {
                        const c = `hsl(${cs.h}, ${cs.s}%, ${cs.l}%)`;
                        values[prop.key] = c;
                        onPropertyChange(prop.key, c);
                      }}
                    />
                  </div>
                  <div class="space-y-0.5">
                    <InputLabel label="Saturation" id="sat-{prop.key}" />
                    <ColorSlider
                      mode="saturation"
                      bind:value={cs.s}
                      hue={cs.h}
                      color={hslString(cs)}
                      onChange={() => {
                        const c = `hsl(${cs.h}, ${cs.s}%, ${cs.l}%)`;
                        values[prop.key] = c;
                        onPropertyChange(prop.key, c);
                      }}
                    />
                  </div>
                  <div class="space-y-0.5">
                    <InputLabel label="Lightness" id="light-{prop.key}" />
                    <ColorSlider
                      mode="lightness"
                      bind:value={cs.l}
                      hue={cs.h}
                      color={hslString(cs)}
                      onChange={() => {
                        const c = `hsl(${cs.h}, ${cs.s}%, ${cs.l}%)`;
                        values[prop.key] = c;
                        onPropertyChange(prop.key, c);
                      }}
                    />
                  </div>
                </Popover.Content>
              </Popover.Root>

              <span class="prop-label">{prop.label}</span>

              <input
                class="hex-input"
                class:text-red-500={colorValue && !isValidColor(colorValue)}
                value={colorValue}
                placeholder=""
                onkeydown={(e) => {
                  if (e.key === "Enter") {
                    e.currentTarget.blur();
                  }
                }}
                onblur={(e) => {
                  const v = e.currentTarget.value;
                  if (v) {
                    values[prop.key] = v;
                    onPropertyChange(prop.key, v);
                    updateColorState(prop.key, v);
                  }
                }}
              />
            </div>
          {/each}
        </div>
      {/if}
    </div>
  {/each}
</div>

<style lang="postcss">
  .sections {
    @apply flex flex-col gap-y-1;
  }

  .properties {
    @apply flex flex-col py-1 px-0.5;
  }

  .property-row {
    @apply flex items-center gap-x-2.5 py-1.5 px-1 pr-2 flex-wrap;
  }

  .swatch {
    @apply w-5 h-5 rounded flex-none border-0 cursor-pointer;
    @apply ring-1 ring-black/10;
  }

  .swatch:hover {
    @apply ring-2 ring-black/20;
  }

  .swatch-empty {
    @apply bg-gray-200;
  }

  .prop-label {
    @apply text-xs text-fg-primary flex-shrink-0;
    width: 80px;
  }

  .hex-input {
    @apply flex-1 min-w-0 text-xs text-right text-fg-muted;
    @apply bg-transparent outline-none;
    @apply border border-gray-200 rounded px-1.5 py-0.5;
  }

  .hex-input:focus {
    @apply text-fg-primary border-primary-400;
  }

  .hex-input:disabled {
    @apply opacity-50;
  }
</style>
