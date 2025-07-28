<script lang="ts">
  import ColorInput from "@rilldata/web-common/components/color-picker/ColorInput.svelte";
  import type { FieldConfig } from "@rilldata/web-common/features/canvas/components/charts/types";
  import type { ChartFieldInput } from "@rilldata/web-common/features/canvas/inspector/types";
  import { COMPARIONS_COLORS } from "@rilldata/web-common/features/dashboards/config";
  import { ChevronDown, ChevronRight } from "lucide-svelte";

  export let fieldConfig: FieldConfig;
  export let onChange: (property: keyof FieldConfig, value: any) => void;
  export let colorMapConfig: ChartFieldInput["colorMappingSelector"];

  let isExpanded = true;

  $: colorValues = colorMapConfig?.values || [];

  $: currentColorMapping = fieldConfig?.colorMapping || [];

  // Initialize color mapping with default colors if not already set
  $: if (colorValues.length > 0 && currentColorMapping.length === 0) {
    currentColorMapping = colorValues.map((value, index) => ({
      value,
      color: COMPARIONS_COLORS[index % COMPARIONS_COLORS.length],
    }));
  }

  // Update color mapping when values change but preserve existing custom colors
  $: if (colorValues.length > 0 && currentColorMapping.length > 0) {
    const existingMapping = new Map(
      currentColorMapping.map((item) => [item.value, item.color]),
    );

    const updatedMapping = colorValues.map((value, index) => ({
      value,
      color:
        existingMapping.get(value) ||
        COMPARIONS_COLORS[index % COMPARIONS_COLORS.length],
    }));

    // Only update if the mapping actually changed
    const hasChanged =
      updatedMapping.length !== currentColorMapping.length ||
      updatedMapping.some(
        (item, index) =>
          item.value !== currentColorMapping[index]?.value ||
          item.color !== currentColorMapping[index]?.color,
      );

    if (hasChanged) {
      onChange("colorMapping", updatedMapping);
    }
  }

  function handleColorChange(value: string, newColor: string) {
    const updatedMapping = currentColorMapping.map((item) =>
      item.value === value ? { ...item, color: newColor } : item,
    );
    onChange("colorMapping", updatedMapping);
  }

  function resetToDefault() {
    const defaultMapping = colorValues.map((value, index) => ({
      value,
      color: COMPARIONS_COLORS[index % COMPARIONS_COLORS.length],
    }));
    onChange("colorMapping", defaultMapping);
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
          <button
            class="text-xs text-blue-600 hover:text-blue-800"
            on:click|stopPropagation={resetToDefault}
          >
            Reset to default
          </button>
        {/if}
        {#if isExpanded}
          <ChevronDown size="14px" class="text-gray-400" />
        {:else}
          <ChevronRight size="14px" class="text-gray-400" />
        {/if}
      </div>
    </button>

    {#if isExpanded}
      <div class="px-1 py-2 max-h-[300px] overflow-y-auto space-y-1">
        {#each currentColorMapping as { value, color } (value)}
          <ColorInput
            small
            stringColor={color}
            labelFirst
            allowLightnessControl
            label={value}
            onChange={(newColor) => handleColorChange(value, newColor)}
          />
        {/each}
        {#if currentColorMapping.length === 0}
          <div class="px-2 py-2 text-xs text-gray-500">
            No color values found
          </div>
        {/if}
      </div>
    {/if}
  </div>
{/if}
