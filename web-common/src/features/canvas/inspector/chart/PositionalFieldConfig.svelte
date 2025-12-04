<script lang="ts">
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import ColorPaletteSelector from "@rilldata/web-common/features/canvas/inspector/chart/field-config/ColorPaletteSelector.svelte";
  import ColorRangeSelector from "@rilldata/web-common/features/canvas/inspector/chart/field-config/ColorRangeSelector.svelte";
  import MultiPositionalFieldsInput from "@rilldata/web-common/features/canvas/inspector/fields/MultiPositionalFieldsInput.svelte";
  import SingleFieldInput from "@rilldata/web-common/features/canvas/inspector/fields/SingleFieldInput.svelte";
  import type { ComponentInputParam } from "@rilldata/web-common/features/canvas/inspector/types";
  import { shouldShowPopover } from "@rilldata/web-common/features/canvas/inspector/util";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import type { FieldConfig } from "@rilldata/web-common/features/components/charts/types";
  import { isFieldConfig } from "@rilldata/web-common/features/components/charts/util";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import FieldConfigPopover from "./field-config/FieldConfigPopover.svelte";
  import MarkTypeToggle from "./field-config/MarkTypeToggle.svelte";

  export let key: string;
  export let config: ComponentInputParam;
  export let metricsView: string;
  export let fieldConfig: FieldConfig;
  export let canvasName: string;

  export let onChange: (updatedConfig: FieldConfig) => void;

  $: ({ instanceId } = $runtime);
  $: ({
    canvasEntity: {
      selectedComponent,
      metricsView: { getTimeDimensionForMetricView },
    },
  } = getCanvasStore(canvasName, instanceId));

  $: chartFieldInput = config.meta?.chartFieldInput;
  $: multiMetricSelector = chartFieldInput?.multiFieldSelector;
  $: colorMapConfig = chartFieldInput?.colorMappingSelector;
  $: colorRangeConfig = chartFieldInput?.colorRangeSelector;

  $: isDimension = chartFieldInput?.type === "dimension";
  $: hasMultipleMeasures = fieldConfig.fields && fieldConfig.fields.length;

  $: timeDimension = getTimeDimensionForMetricView(metricsView);

  function updateFieldConfig(fieldName: string) {
    const isTime = $timeDimension && fieldName === $timeDimension;

    let updatedConfig: FieldConfig;
    if (isTime && $timeDimension) {
      updatedConfig = {
        ...fieldConfig,
        field: $timeDimension,
        type: "temporal",
        sort: undefined,
      };
    } else {
      updatedConfig = {
        ...fieldConfig,
        field: fieldName,
        type: isTime ? "temporal" : isDimension ? "nominal" : "quantitative",
        sort: undefined,
      };
    }

    onChange(updatedConfig);
  }

  function updateFieldProperty(property: keyof FieldConfig, value: any) {
    if (fieldConfig[property] === value) {
      return;
    }

    const updatedConfig: FieldConfig = {
      ...fieldConfig,
      [property]: value,
    };

    if (property === "limit" && Array.isArray(updatedConfig.sort)) {
      updatedConfig.sort = "-x";
    }

    onChange(updatedConfig);
  }

  function handleMultiFieldUpdate(items: string[]) {
    // Handle transitions between single and multi-measure modes
    const currentMultiMeasures = fieldConfig.fields || [];
    const updatedMultiMeasures = items;

    let updatedConfig: FieldConfig = { ...fieldConfig };

    // Transition from single to multi-measure mode
    if (
      currentMultiMeasures.length === 0 &&
      updatedMultiMeasures &&
      updatedMultiMeasures.length > 0 &&
      fieldConfig.field
    ) {
      const measuresSet = new Set([fieldConfig.field, ...updatedMultiMeasures]);
      updatedConfig = {
        ...fieldConfig,
        fields: Array.from(measuresSet),
      };
    }
    // Transition from multi to single-measure mode
    else if (
      currentMultiMeasures.length > 1 &&
      updatedMultiMeasures &&
      updatedMultiMeasures.length === 1
    ) {
      // When down to one measure, move it to the main field and clear fields array
      const singleMeasure = updatedMultiMeasures[0];
      updatedConfig = {
        ...fieldConfig,
        field: singleMeasure,
        fields: undefined,
      };
    }
    // Normal multi-field update
    else {
      updatedConfig = {
        ...fieldConfig,
        fields: items,
      };
    }

    onChange(updatedConfig);
  }

  $: popoverKey = `${$selectedComponent}-${metricsView}-${fieldConfig.field}`;
  $: hasPopoverContent = shouldShowPopover(chartFieldInput);
</script>

<div class="gap-y-1">
  <div class="flex justify-between items-center">
    <InputLabel small label={config.label ?? key} id={key} />
    {#key popoverKey}
      {#if hasPopoverContent}
        <FieldConfigPopover
          {fieldConfig}
          label={config.label ?? key}
          onChange={updateFieldProperty}
          {chartFieldInput}
        />
      {/if}
    {/key}
  </div>

  <div class="flex flex-col gap-y-2">
    {#if !hasMultipleMeasures}
      <SingleFieldInput
        {canvasName}
        metricName={metricsView}
        id={`${key}-field`}
        type={isDimension ? "dimension" : "measure"}
        includeTime={!chartFieldInput?.hideTimeDimension}
        excludedValues={chartFieldInput?.excludedValues}
        selectedItem={fieldConfig?.field}
        onSelect={async (field) => {
          updateFieldConfig(field);
        }}
      />
      {#if isFieldConfig(fieldConfig) && colorMapConfig?.enable}
        <div class="pt-2">
          <ColorPaletteSelector
            colorMapping={fieldConfig.colorMapping}
            onChange={updateFieldProperty}
            {colorMapConfig}
          />
        </div>
      {/if}
      {#if isFieldConfig(fieldConfig) && colorRangeConfig?.enable}
        <div class="pt-2">
          <ColorRangeSelector
            colorRange={fieldConfig.colorRange}
            onChange={updateFieldProperty}
            {colorRangeConfig}
            {canvasName}
          />
        </div>
      {/if}
    {/if}
    {#if multiMetricSelector}
      <MultiPositionalFieldsInput
        {canvasName}
        metricName={metricsView}
        selectedItems={fieldConfig.fields?.length
          ? fieldConfig.fields
          : [fieldConfig.field]}
        types={isDimension ? ["dimension"] : ["measure"]}
        excludedValues={chartFieldInput?.excludedValues}
        chipItems={fieldConfig.fields}
        onMultiSelect={handleMultiFieldUpdate}
      />
    {/if}
  </div>

  {#if chartFieldInput?.markTypeSelector}
    <MarkTypeToggle
      selectedMark={fieldConfig.mark}
      onClick={(mark) => {
        updateFieldProperty("mark", mark);
      }}
    />
  {/if}
</div>
