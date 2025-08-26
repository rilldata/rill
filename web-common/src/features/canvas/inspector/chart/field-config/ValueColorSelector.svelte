<script lang="ts">
  import ColorInput from "@rilldata/web-common/components/color-picker/ColorInput.svelte";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import type { ChartFieldInput } from "@rilldata/web-common/features/canvas/inspector/types";
  import { COMPARIONS_COLORS } from "@rilldata/web-common/features/dashboards/config";

  export let label: string;
  export let colorValues: string[] = [];
  export let colorMapping: { value: string; color: string }[] | undefined;
  export let onChange: (
    updatedMapping: { value: string; color: string }[] | undefined,
  ) => void;
  export let colorMapConfig: ChartFieldInput["colorMappingSelector"];

  $: currentColorMapping = colorMapping || [];

  $: allColorMappings = colorValues.map((value, index) => {
    const existingMapping = currentColorMapping.find(
      (item) => item.value === value,
    );
    const defaultColor = COMPARIONS_COLORS[index % COMPARIONS_COLORS.length];
    return {
      value,
      color: existingMapping?.color || defaultColor,
    };
  });

  function handleColorChange(value: string, newColor: string) {
    const valueIndex = colorValues.findIndex((v) => v === value);
    const defaultColor =
      COMPARIONS_COLORS[valueIndex % COMPARIONS_COLORS.length];

    let updatedMapping: { value: string; color: string }[];

    if (newColor === defaultColor) {
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
          index === existingIndex ? { ...item, color: newColor } : item,
        );
      } else {
        updatedMapping = [...currentColorMapping, { value, color: newColor }];
      }
    }

    onChange(updatedMapping.length > 0 ? updatedMapping : undefined);
  }
</script>

{#if colorMapConfig?.enable && colorValues.length > 0}
  <div class="space-y-2">
    <InputLabel small {label} id="color-mapping-label" />
    <div class="space-y-1">
      {#each allColorMappings as { value, color } (value)}
        <ColorInput
          small
          stringColor={color}
          labelFirst
          allowLightnessControl
          label={value}
          onChange={(newColor) => handleColorChange(value, newColor)}
        />
      {/each}
    </div>
  </div>
{/if}
