<script lang="ts">
  import FieldSwitcher from "@rilldata/web-common/components/forms/FieldSwitcher.svelte";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import SingleFieldInput from "@rilldata/web-common/features/canvas/inspector/fields/SingleFieldInput.svelte";
  import type { ComponentInputParam } from "@rilldata/web-common/features/canvas/inspector/types";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import { type FieldConfig } from "@rilldata/web-common/features/components/charts/types";
  import { isFieldConfig } from "@rilldata/web-common/features/components/charts/util";
  import { themeControl } from "@rilldata/web-common/features/themes/theme-control";
  import { resolveThemeColors } from "@rilldata/web-common/features/themes/theme-utils";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import ColorPaletteSelector from "./field-config/ColorPaletteSelector.svelte";
  import FieldConfigPopover from "./field-config/FieldConfigPopover.svelte";
  import SingleColorSelector from "./field-config/SingleColorSelector.svelte";

  export let key: string;
  export let metricsView: string;
  export let markConfig: FieldConfig | string;
  export let config: ComponentInputParam;
  export let canvasName: string;
  export let onChange: (updatedConfig: FieldConfig | string) => void;

  $: ({ instanceId } = $runtime);
  $: ({
    canvasEntity: { selectedComponent, theme },
  } = getCanvasStore(canvasName, instanceId));

  $: isThemeModeDark = $themeControl === "dark";
  $: resolvedTheme = resolveThemeColors($theme?.spec, isThemeModeDark);

  $: selected = !markConfig || typeof markConfig === "string" ? 0 : 1;

  $: chartFieldInput = config.meta?.chartFieldInput;
  $: colorMapConfig = chartFieldInput?.colorMappingSelector;

  $: isValue = chartFieldInput?.type === "value";

  function updateFieldConfig(property: keyof FieldConfig, value: any) {
    if (typeof markConfig !== "string") {
      if (markConfig[property] === value) {
        return;
      }
      const updatedConfig: FieldConfig = {
        ...markConfig,
        [property]: value,
      };

      if (property === "field") {
        updatedConfig.colorMapping = undefined;
      }

      onChange(updatedConfig);
    } else if (property === "field") {
      // switch to field from single color
      onChange({
        field: value,
        type: "nominal",
      });
    }
  }

  $: popoverKey = `${$selectedComponent}-${metricsView}-${typeof markConfig === "string" ? markConfig : markConfig?.field}`;
</script>

<div class="space-y-2">
  <div class="flex justify-between items-center">
    <InputLabel small label={config.label ?? key} id={key} />
    {#if Object.keys(chartFieldInput ?? {}).length > 1 && typeof markConfig !== "string"}
      {#key popoverKey}
        <FieldConfigPopover
          fieldConfig={markConfig}
          label={config.label ?? key}
          onChange={updateFieldConfig}
          {chartFieldInput}
        />
      {/key}
    {/if}
  </div>

  {#if !isValue}
    <FieldSwitcher
      small
      fields={["One color", "Split by"]}
      {selected}
      onClick={(_, field) => {
        if (field === "One color") {
          selected = 0;
          onChange(typeof markConfig === "string" ? markConfig : "primary");
        } else if (field === "Split by") {
          selected = 1;
        }
      }}
    />
  {/if}
</div>

{#if isValue && colorMapConfig?.enable && typeof markConfig === "object"}
  <div class="pt-2">
    <ColorPaletteSelector
      colorMapping={markConfig?.colorMapping}
      onChange={updateFieldConfig}
      {colorMapConfig}
    />
  </div>
{:else if selected === 0}
  <div class="pt-2">
    {#key `${isThemeModeDark}-${resolvedTheme.primary.hex()}-${resolvedTheme.secondary.hex()}`}
      <SingleColorSelector
        small
        theme={resolvedTheme}
        markConfig={typeof markConfig === "string" ? markConfig : "primary"}
        onChange={(newColor) => {
          onChange(newColor);
        }}
      />
    {/key}
  </div>
{:else if selected === 1}
  <SingleFieldInput
    {canvasName}
    metricName={metricsView}
    id={`${key}-field`}
    type="dimension"
    excludedValues={chartFieldInput?.excludedValues}
    selectedItem={typeof markConfig === "string"
      ? undefined
      : markConfig?.field}
    onSelect={async (field) => {
      updateFieldConfig("field", field);
    }}
  />

  {#if isFieldConfig(markConfig) && colorMapConfig?.enable}
    <div class="pt-2">
      <ColorPaletteSelector
        colorMapping={markConfig?.colorMapping}
        onChange={updateFieldConfig}
        {colorMapConfig}
      />
    </div>
  {/if}
{/if}
