<!--
  ColorRangeSelector - Component for configuring continuous color ranges in charts like heatmaps.
  Supports both predefined color schemes (tealblues, magma, etc.) and custom color gradients.
  Defaults to tealblues scheme but allows switching to custom gradient mode.
-->
<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import * as m from "@rilldata/web-common/paraglide/messages.js";
  import ColorInput from "@rilldata/web-common/components/color-picker/ColorInput.svelte";
  import FieldSwitcher from "@rilldata/web-common/components/forms/FieldSwitcher.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import type { ChartFieldInput } from "@rilldata/web-common/features/canvas/inspector/types";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import type {
    ColorRangeMapping,
    FieldConfig,
  } from "@rilldata/web-common/features/components/charts/types";
  import {
    defaultPrimaryColors,
    defaultSecondaryColors,
  } from "@rilldata/web-common/features/themes/color-config";
  import { themeControl } from "@rilldata/web-common/features/themes/theme-control";
  import { resolveThemeColors } from "@rilldata/web-common/features/themes/theme-utils";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { slide } from "svelte/transition";
  import type { ColorScheme } from "vega";

  export let colorRange: ColorRangeMapping | undefined;
  export let onChange: (property: keyof FieldConfig, value: any) => void;
  export let colorRangeConfig: ChartFieldInput["colorRangeSelector"];
  export let canvasName: string;

  const client = useRuntimeClient();

  // Available Vega-Lite color schemes
  // https://vega.github.io/vega/docs/schemes/
  $: colorSchemes = [
    { label: m.canvas_sequential_theme(), value: "sequential" as const },
    { label: m.canvas_diverging_theme(), value: "diverging" as const },
    { label: m.canvas_teal_blues(), value: "tealblues" as ColorScheme },
    { label: m.canvas_viridis(), value: "viridis" as ColorScheme },
    { label: m.canvas_magma(), value: "magma" as ColorScheme },
    { label: m.canvas_inferno(), value: "inferno" as ColorScheme },
    { label: m.canvas_plasma(), value: "plasma" as ColorScheme },
    { label: m.canvas_cividis(), value: "cividis" as ColorScheme },
    { label: m.canvas_blues(), value: "blues" as ColorScheme },
    { label: m.canvas_teals(), value: "teals" as ColorScheme },
    { label: m.canvas_greens(), value: "greens" as ColorScheme },
    { label: m.canvas_greys(), value: "greys" as ColorScheme },
    { label: m.canvas_oranges(), value: "oranges" as ColorScheme },
    { label: m.canvas_purples(), value: "purples" as ColorScheme },
    { label: m.canvas_reds(), value: "reds" as ColorScheme },
    { label: m.canvas_turbo(), value: "turbo" as ColorScheme },
    { label: m.canvas_spectral(), value: "spectral" as ColorScheme },
  ] as { label: string; value: ColorScheme | "sequential" | "diverging" }[];

  $: ({
    canvasEntity: { theme },
  } = getCanvasStore(canvasName, client.instanceId));

  $: isThemeModeDark = $themeControl === "dark";
  $: resolvedTheme = resolveThemeColors($theme?.spec, isThemeModeDark);

  $: currentColorRange =
    colorRange ||
    ({
      mode: "scheme",
      scheme: "tealblues",
    } as ColorRangeMapping);

  $: currentMode = currentColorRange.mode || "scheme";

  function isGradientMode(
    colorRange: ColorRangeMapping,
  ): colorRange is { mode: "gradient"; start: string; end: string } {
    return colorRange.mode === "gradient";
  }

  $: resolveColor = (color: string): string => {
    if (color === "primary") {
      return (
        resolvedTheme.primary?.css("hsl") || `hsl(${defaultPrimaryColors[500]})`
      );
    } else if (color === "secondary") {
      return (
        resolvedTheme.secondary?.css("hsl") ||
        `hsl(${defaultSecondaryColors[500]})`
      );
    }
    return color;
  };

  function handleModeSwitch(mode: "scheme" | "gradient") {
    let updatedRange: ColorRangeMapping;

    if (mode === "scheme") {
      updatedRange = {
        mode: "scheme",
        scheme: "tealblues",
      };
    } else {
      updatedRange = {
        mode: "gradient",
        start: "primary",
        end: "secondary",
      };
    }

    onChange("colorRange", updatedRange);
  }

  function handleSchemeChange(
    scheme: ColorScheme | "sequential" | "diverging",
  ) {
    const updatedRange: ColorRangeMapping = {
      mode: "scheme",
      scheme,
    };
    onChange("colorRange", updatedRange);
  }

  function handleStartColorChange(newColor: string) {
    let currentEnd = "secondary";

    if (isGradientMode(currentColorRange)) {
      currentEnd = currentColorRange.end;
    }

    const updatedRange: ColorRangeMapping = {
      mode: "gradient",
      start: newColor,
      end: currentEnd,
    };
    onChange("colorRange", updatedRange);
  }

  function handleEndColorChange(newColor: string) {
    let currentStart = "primary";

    if (isGradientMode(currentColorRange)) {
      currentStart = currentColorRange.start;
    }

    const updatedRange: ColorRangeMapping = {
      mode: "gradient",
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
        fields={[m.canvas_scheme(), m.canvas_gradient()]}
        selected={currentMode === "scheme" ? 0 : 1}
        onClick={(i) => handleModeSwitch(i === 0 ? "scheme" : "gradient")}
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
        <!-- Custom Gradient Selectors -->
        {#key `${isThemeModeDark}-${resolvedTheme.primary.hex()}`}
          <ColorInput
            small
            stringColor={resolveColor(
              isGradientMode(currentColorRange)
                ? currentColorRange.start
                : "primary",
            )}
            labelFirst
            allowLightnessControl
            label={m.canvas_start_color()}
            onChange={handleStartColorChange}
          />
        {/key}

        {#key `${isThemeModeDark}-${resolvedTheme.secondary.hex()}`}
          <ColorInput
            small
            stringColor={resolveColor(
              isGradientMode(currentColorRange)
                ? currentColorRange.end
                : "secondary",
            )}
            labelFirst
            allowLightnessControl
            label={m.canvas_end_color()}
            onChange={handleEndColorChange}
          />
        {/key}
      {/if}

      <div class="px-1 flex items-center justify-end">
        <Button type="text" onClick={resetToDefault}
          >{m.canvas_reset_to_default()}</Button
        >
      </div>
    </div>
  </div>
{/if}
