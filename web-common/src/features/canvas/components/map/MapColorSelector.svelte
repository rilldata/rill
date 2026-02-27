<script lang="ts">
  import FieldSwitcher from "@rilldata/web-common/components/forms/FieldSwitcher.svelte";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import SingleFieldInput from "@rilldata/web-common/features/canvas/inspector/fields/SingleFieldInput.svelte";
  import ColorRangeSelector from "@rilldata/web-common/features/canvas/inspector/chart/field-config/ColorRangeSelector.svelte";
  import SingleColorSelector from "@rilldata/web-common/features/canvas/inspector/chart/field-config/SingleColorSelector.svelte";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import { themeControl } from "@rilldata/web-common/features/themes/theme-control";
  import { resolveThemeColors } from "@rilldata/web-common/features/themes/theme-utils";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import type { MapColorConfig } from ".";
  import type { ColorRangeMapping } from "@rilldata/web-common/features/components/charts/types";

  export let canvasName: string;
  export let metricsView: string;
  export let colorConfig: string | MapColorConfig;
  export let onChange: (updatedConfig: string | MapColorConfig) => void;

  $: ({ instanceId } = $runtime);
  $: ({
    canvasEntity: { theme },
  } = getCanvasStore(canvasName, instanceId));

  $: isThemeModeDark = $themeControl === "dark";
  $: resolvedTheme = resolveThemeColors($theme?.spec, isThemeModeDark);

  // 0 = One color, 1 = By measure
  $: selected = typeof colorConfig === "object" && colorConfig !== null ? 1 : 0;

  $: currentColorRange =
    typeof colorConfig === "object" ? colorConfig.colorRange : undefined;

  $: currentMeasure =
    typeof colorConfig === "object" ? colorConfig.measure : undefined;

  function handleModeSwitch(_: number, field: string) {
    if (field === "One color") {
      onChange(typeof colorConfig === "string" ? colorConfig : "primary");
    } else {
      // Switch to by-measure mode, preserve existing measure if any
      onChange({
        measure: currentMeasure ?? "",
        colorRange: currentColorRange ?? {
          mode: "scheme",
          scheme: "tealblues",
        },
      });
    }
  }

  function handleMeasureSelect(measure: string) {
    const range =
      currentColorRange ?? ({ mode: "scheme", scheme: "tealblues" } as const);
    onChange({ measure, colorRange: range });
  }

  function handleColorRangeChange(_property: string, value: ColorRangeMapping) {
    onChange({
      measure: currentMeasure ?? "",
      colorRange: value,
    });
  }

  function handleSingleColorChange(newColor: string) {
    onChange(newColor);
  }
</script>

<div class="space-y-2">
  <InputLabel small label="Color" id="map-color" />

  <FieldSwitcher
    small
    fields={["One color", "By measure"]}
    {selected}
    onClick={handleModeSwitch}
  />
</div>

{#if selected === 0}
  <div class="pt-2">
    {#key `${isThemeModeDark}-${resolvedTheme.primary.hex()}-${resolvedTheme.secondary.hex()}`}
      <SingleColorSelector
        small
        theme={resolvedTheme}
        markConfig={typeof colorConfig === "string" ? colorConfig : "primary"}
        onChange={handleSingleColorChange}
      />
    {/key}
  </div>
{:else}
  <div class="pt-2 space-y-2">
    <SingleFieldInput
      {canvasName}
      metricName={metricsView}
      id="map-color-measure"
      type="measure"
      selectedItem={currentMeasure}
      onSelect={handleMeasureSelect}
    />

    <ColorRangeSelector
      {canvasName}
      colorRange={currentColorRange}
      onChange={handleColorRangeChange}
      colorRangeConfig={{ enable: true }}
    />
  </div>
{/if}
