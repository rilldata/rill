<script lang="ts">
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import ColorInput from "@rilldata/web-common/components/color-picker/ColorInput.svelte";
  import * as Popover from "@rilldata/web-common/components/popover";
  import type { FieldConfig } from "@rilldata/web-common/features/canvas/components/charts/types";
  import type { ChartFieldInput } from "@rilldata/web-common/features/canvas/inspector/types";
  import { COMPARIONS_COLORS } from "@rilldata/web-common/features/dashboards/config";
  import { Palette } from "lucide-svelte";

  export let fieldConfig: FieldConfig;
  export let onChange: (property: keyof FieldConfig, value: any) => void;
  export let colorMapConfig: ChartFieldInput["colorMappingSelector"];

  let isColorMappingDropdownOpen = false;

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
</script>

{#if colorMapConfig?.enable && colorValues.length > 0}
  <div class="py-1 flex items-center justify-between">
    <span class="text-xs">Color mapping</span>
    <div class="flex items-center gap-x-1">
      <Popover.Root bind:open={isColorMappingDropdownOpen}>
        <Popover.Trigger>
          <IconButton rounded active={isColorMappingDropdownOpen}>
            <Palette size="14px" />
          </IconButton>
        </Popover.Trigger>
        <Popover.Content align="end" class="w-[280px] p-0">
          <div
            class="px-3 py-2 border-b border-gray-200 flex items-center justify-between"
          >
            <span class="text-xs font-medium">Color Mapping</span>
            <button
              class="text-xs text-blue-600 hover:text-blue-800"
              on:click={resetToDefault}
            >
              Reset to default
            </button>
          </div>
          <div class="px-3 py-2 max-h-[300px] overflow-y-auto space-y-2">
            {#each currentColorMapping as { value, color } (value)}
              <div class="flex items-center gap-x-3">
                <div class="flex-1 min-w-0">
                  <span class="text-xs truncate block" title={value}
                    >{value}</span
                  >
                </div>
                <div class="flex-none">
                  <ColorInput
                    small
                    stringColor={color}
                    showLabel={false}
                    label=""
                    onChange={(newColor) => handleColorChange(value, newColor)}
                  />
                </div>
              </div>
            {/each}
            {#if currentColorMapping.length === 0}
              <div class="px-2 py-2 text-xs text-gray-500">
                No color values found
              </div>
            {/if}
          </div>
        </Popover.Content>
      </Popover.Root>
    </div>
  </div>
{/if}
