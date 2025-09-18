<!--
  ColorRangeSelector - Component for configuring continuous color ranges in charts like heatmaps.
  Supports both predefined color schemes (tealblues, magma, etc.) and custom color ranges.
  Defaults to tealblues scheme but allows switching to custom range mode.
-->
<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import ColorInput from "@rilldata/web-common/components/color-picker/ColorInput.svelte";
  import FieldSwitcher from "@rilldata/web-common/components/forms/FieldSwitcher.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import type { FieldConfig } from "@rilldata/web-common/features/canvas/components/charts/types";
  import type {
    ChartFieldInput,
    ColorRangeMapping,
  } from "@rilldata/web-common/features/canvas/inspector/types";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import {
    defaultPrimaryColors,
    defaultSecondaryColors,
  } from "@rilldata/web-common/features/themes/color-config";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { slide } from "svelte/transition";
  import type { ColorScheme } from "vega-typings";

  export let colorRange: ColorRangeMapping | undefined;
  export let onChange: (property: keyof FieldConfig, value: any) => void;
  export let colorRangeConfig: ChartFieldInput["colorRangeSelector"];
  export let canvasName: string;

  // Available Vega-Lite color schemes
  // https://vega.github.io/vega/docs/schemes/
  const colorSchemes: { label: string; value: ColorScheme }[] = [
    { label: "Teal blues", value: "tealblues" },
    { label: "Viridis", value: "viridis" },
    { label: "Magma", value: "magma" },
    { label: "Inferno", value: "inferno" },
    { label: "Plasma", value: "plasma" },
    { label: "Cividis", value: "cividis" },
    { label: "Blues", value: "blues" },
    { label: "Teals", value: "teals" },
    { label: "Greens", value: "greens" },
    { label: "Greys", value: "greys" },
    { label: "Oranges", value: "oranges" },
    { label: "Purples", value: "purples" },
    { label: "Reds", value: "reds" },
    { label: "Turbo", value: "turbo" },
    { label: "Spectral", value: "spectral" },
  ];

  $: ({ instanceId } = $runtime);
  $: ({
    canvasEntity: { theme },
  } = getCanvasStore(canvasName, instanceId));

  $: currentColorRange =
    colorRange ||
    ({
      mode: "scheme",
      scheme: "tealblues",
    } as ColorRangeMapping);

  $: currentMode = currentColorRange.mode || "scheme";

  function isRangeMode(
    colorRange: ColorRangeMapping,
  ): colorRange is { mode: "range"; start: string; end: string } {
    return colorRange.mode === "range";
  }

  const resolveColor = (color: string): string => {
    if (color === "primary") {
      return $theme.primary?.css("hsl") || `hsl(${defaultPrimaryColors[500]})`;
    } else if (color === "secondary") {
      return (
        $theme.secondary?.css("hsl") || `hsl(${defaultSecondaryColors[500]})`
      );
    }
    return color;
  };

  const getColorLabel = (color: string): string => {
    switch (color) {
      case "primary":
        return "Primary";
      case "secondary":
        return "Secondary";
      default:
        return "";
    }
  };

  function handleModeSwitch(mode: "scheme" | "range") {
    let updatedRange: ColorRangeMapping;

    if (mode === "scheme") {
      updatedRange = {
        mode: "scheme",
        scheme: "tealblues",
      };
    } else {
      updatedRange = {
        mode: "range",
        start: "primary",
        end: "secondary",
      };
    }

    onChange("colorRange", updatedRange);
  }

  function handleSchemeChange(scheme: ColorScheme) {
    const updatedRange: ColorRangeMapping = {
      mode: "scheme",
      scheme,
    };
    onChange("colorRange", updatedRange);
  }

  function handleStartColorChange(newColor: string) {
    let currentEnd = "secondary";

    if (isRangeMode(currentColorRange)) {
      currentEnd = currentColorRange.end;
    }

    const updatedRange: ColorRangeMapping = {
      mode: "range",
      start: newColor,
      end: currentEnd,
    };
    onChange("colorRange", updatedRange);
  }

  function handleEndColorChange(newColor: string) {
    let currentStart = "primary";

    if (isRangeMode(currentColorRange)) {
      currentStart = currentColorRange.start;
    }

    const updatedRange: ColorRangeMapping = {
      mode: "range",
      start: currentStart,
      end: newColor,
    };
    onChange("colorRange", updatedRange);
  }

  function resetToDefault() {
    onChange("colorRange", {
      mode: "scheme",
      scheme: "tealblues",
    });
  }
</script>

{#if colorRangeConfig?.enable}
  <div>
    <div class="space-y-2" transition:slide={{ duration: 200 }}>
      <!-- Mode Switcher -->
      <FieldSwitcher
        small
        fields={["Scheme", "Range"]}
        selected={currentMode === "scheme" ? 0 : 1}
        onClick={(i, value) =>
          handleModeSwitch(value === "Scheme" ? "scheme" : "range")}
      />

      {#if currentMode === "scheme"}
        <!-- Color Scheme Selector -->
        <Select
          size="sm"
          sameWidth
          id="color-scheme-select"
          options={colorSchemes}
          value={currentColorRange.mode === "scheme"
            ? currentColorRange.scheme
            : "tealblues"}
          onChange={handleSchemeChange}
        />
      {:else}
        <!-- Custom Range Selectors -->
        <ColorInput
          small
          stringColor={resolveColor(
            isRangeMode(currentColorRange)
              ? currentColorRange.start
              : "primary",
          )}
          labelFirst
          allowLightnessControl
          label="Start color"
          onChange={handleStartColorChange}
        />

        <ColorInput
          small
          stringColor={resolveColor(
            isRangeMode(currentColorRange)
              ? currentColorRange.end
              : "secondary",
          )}
          labelFirst
          allowLightnessControl
          label="End color"
          onChange={handleEndColorChange}
        />
      {/if}

      <div class="px-1 flex items-center justify-end">
        <Button type="text" onClick={resetToDefault}>Reset to default</Button>
      </div>
    </div>
  </div>
{/if}
