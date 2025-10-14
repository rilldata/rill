<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import ColorInput from "@rilldata/web-common/components/color-picker/ColorInput.svelte";
  import type { ChartFieldInput } from "@rilldata/web-common/features/canvas/inspector/types";
  import type {
    ColorMapping,
    FieldConfig,
  } from "@rilldata/web-common/features/components/charts/types";
  import {
    colorToVariableReference,
    getColorForValues,
    resolveCSSVariable,
  } from "@rilldata/web-common/features/components/charts/util";
  import { COMPARIONS_COLORS } from "@rilldata/web-common/features/dashboards/config";
  import { ChevronDown, ChevronRight } from "lucide-svelte";
  import { slide } from "svelte/transition";

  export let colorMapping: ColorMapping | undefined;
  export let onChange: (property: keyof FieldConfig, value: any) => void;
  export let colorMapConfig: ChartFieldInput["colorMappingSelector"];

  const THRESHOLD = 11;

  let isExpanded = true;
  let showAllValues = false;

  $: colorValues = colorMapConfig?.values || [];

  $: currentColorMapping = colorMapping || [];

  $: allColorMappings =
    getColorForValues(colorValues, currentColorMapping) || [];

  $: displayedColorMappings = showAllValues
    ? allColorMappings
    : allColorMappings.slice(0, THRESHOLD);

  $: hasMoreThanThreshold = allColorMappings.length > THRESHOLD;

  function handleColorChange(value: string, newColor: string) {
    const valueIndex = colorValues.findIndex((v) => v === value);
    const defaultColorVar =
      COMPARIONS_COLORS[valueIndex % COMPARIONS_COLORS.length];

    // Convert the color back to a CSS variable reference if it matches a palette color
    const colorToSave = colorToVariableReference(newColor);

    let updatedMapping: ColorMapping;

    if (colorToSave === defaultColorVar) {
      // Remove from custom mappings if it's set back to default
      updatedMapping = currentColorMapping.filter(
        (item) => item.value !== value,
      );
    } else {
      // Add or update custom mapping
      const existingIndex = currentColorMapping.findIndex(
        (item) => item.value === value,
      );
      if (existingIndex >= 0) {
        updatedMapping = currentColorMapping.map((item, index) =>
          index === existingIndex ? { ...item, color: colorToSave } : item,
        );
      } else {
        updatedMapping = [
          ...currentColorMapping,
          { value, color: colorToSave },
        ];
      }
    }

    onChange(
      "colorMapping",
      updatedMapping.length > 0 ? updatedMapping : undefined,
    );
  }

  function resetToDefault() {
    onChange("colorMapping", undefined);
  }

  function toggleExpanded() {
    isExpanded = !isExpanded;
  }
</script>

{#if colorMapConfig?.enable && colorValues.length > 0}
  <div>
    <button
      class="w-full p-1 flex items-center justify-between hover:bg-gray-50"
      on:click={toggleExpanded}
    >
      <span class="text-xs font-medium">Color mapping</span>
      <div class="flex items-center gap-x-2">
        {#if isExpanded}
          <ChevronDown size="14px" class="text-gray-400" />
        {:else}
          <ChevronRight size="14px" class="text-gray-400" />
        {/if}
      </div>
    </button>

    {#if isExpanded}
      <div
        class="px-1 py-2 overflow-y-auto space-y-1"
        transition:slide={{ duration: 200 }}
      >
        {#each displayedColorMappings as { value, color } (value)}
          <ColorInput
            small
            stringColor={resolveCSSVariable(color)}
            labelFirst
            allowLightnessControl
            label={value}
            onChange={(newColor) => handleColorChange(value, newColor)}
          />
        {/each}
        {#if allColorMappings.length === 0}
          <div class="px-2 py-2 text-xs text-gray-500">
            No color values found
          </div>
        {/if}
        <div class="p-1 flex items-center justify-between">
          <div>
            {#if hasMoreThanThreshold && !showAllValues}
              <Button type="text" onClick={() => (showAllValues = true)}>
                See {allColorMappings.length - THRESHOLD} more value{allColorMappings.length -
                  THRESHOLD !==
                1
                  ? "s"
                  : ""}
              </Button>
            {:else if hasMoreThanThreshold && showAllValues}
              <Button type="text" onClick={() => (showAllValues = false)}>
                See less
              </Button>
            {/if}
          </div>
          <Button type="text" onClick={resetToDefault}>Reset to default</Button>
        </div>
      </div>
    {/if}
  </div>
{/if}
