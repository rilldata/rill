<script lang="ts">
  import ColorInput from "@rilldata/web-common/components/color-picker/ColorInput.svelte";
  import FieldSwitcher from "@rilldata/web-common/components/forms/FieldSwitcher.svelte";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import type { FieldConfig } from "@rilldata/web-common/features/canvas/components/charts/types";
  import SingleFieldInput from "@rilldata/web-common/features/canvas/inspector/SingleFieldInput.svelte";
  import type { ComponentInputParam } from "@rilldata/web-common/features/canvas/inspector/types";
  import {
    defaultPrimaryColors,
    defaultSecondaryColors,
  } from "@rilldata/web-common/features/themes/color-config";
  import chroma from "chroma-js";
  import FieldConfigPopover from "./field-config/FieldConfigPopover.svelte";

  export let key: string;
  export let metricsView: string;
  export let markConfig: FieldConfig | string;
  export let config: ComponentInputParam;
  export let canvasName: string;
  export let onChange: (updatedConfig: FieldConfig | string) => void;

  $: selected = !markConfig || typeof markConfig === "string" ? 0 : 1;

  const theme = {
    primary: chroma(`hsl(${defaultPrimaryColors[500]})`),
    secondary: chroma(`hsl(${defaultSecondaryColors[500]})`),
  };

  // Get the actual color value based on whether it's a theme color or regular color
  $: color = typeof markConfig === "string" ? markConfig : theme.primary.hex();

  // Get the display color - actual color for rendering
  $: displayColor = (() => {
    if (typeof markConfig !== "string") return theme.primary.hex();

    switch (markConfig) {
      case "primary":
        return theme.primary.hex();
      case "secondary":
        return theme.secondary.hex();
      default:
        return markConfig;
    }
  })();

  // Get the display label for the color input
  $: colorLabel = (() => {
    if (typeof markConfig !== "string") return "";

    switch (markConfig) {
      case "primary":
        return "Primary";
      case "secondary":
        return "Secondary";
      default:
        return markConfig;
    }
  })();

  $: chartFieldInput = config.meta?.chartFieldInput;

  function updateFieldConfig(property: keyof FieldConfig, value: any) {
    if (typeof markConfig !== "string") {
      if (markConfig[property] === value) {
        return;
      }
      const updatedConfig: FieldConfig = {
        ...markConfig,
        [property]: value,
      };
      onChange(updatedConfig);
    } else if (property === "field") {
      onChange({
        field: value,
        type: "nominal",
      });
    }
  }
</script>

<div class="space-y-2">
  <div class="flex justify-between items-center">
    <InputLabel small label={config.label ?? key} id={key} />
    {#if Object.keys(chartFieldInput ?? {}).length > 1 && typeof markConfig !== "string"}
      {#key markConfig}
        <FieldConfigPopover
          fieldConfig={markConfig}
          label={config.label ?? key}
          onChange={updateFieldConfig}
          {chartFieldInput}
        />
      {/key}
    {/if}
  </div>

  <FieldSwitcher
    small
    fields={["One color", "Split by"]}
    {selected}
    onClick={(_, field) => {
      if (field === "One color") {
        selected = 0;
        onChange(color);
      } else if (field === "Split by") {
        selected = 1;
      }
    }}
  />
</div>

{#if selected === 0}
  <div class="pt-2">
    <ColorInput
      small
      stringColor={displayColor}
      label={colorLabel}
      onChange={(newColor) => {
        onChange(newColor);
      }}
    />
  </div>
{:else if selected === 1}
  <SingleFieldInput
    {canvasName}
    metricName={metricsView}
    id={`${key}-field`}
    type="dimension"
    selectedItem={typeof markConfig === "string"
      ? undefined
      : markConfig?.field}
    onSelect={async (field) => {
      updateFieldConfig("field", field);
    }}
  />
{/if}
