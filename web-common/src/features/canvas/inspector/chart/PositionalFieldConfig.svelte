<script lang="ts">
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import type { FieldConfig } from "@rilldata/web-common/features/canvas/components/charts/types";
  import SingleFieldInput from "@rilldata/web-common/features/canvas/inspector/SingleFieldInput.svelte";
  import type { ComponentInputParam } from "@rilldata/web-common/features/canvas/inspector/types";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import FieldConfigPopover from "./field-config/FieldConfigPopover.svelte";

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
      spec: { getTimeDimensionForMetricView },
    },
  } = getCanvasStore(canvasName, instanceId));

  $: chartFieldInput = config.meta?.chartFieldInput;

  $: isDimension = chartFieldInput?.type === "dimension";

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

  $: popoverKey = `${$selectedComponent}-${metricsView}-${fieldConfig.field}`;
</script>

<div class="gap-y-1">
  <div class="flex justify-between items-center">
    <InputLabel small label={config.label ?? key} id={key} />
    {#key popoverKey}
      {#if Object.keys(chartFieldInput ?? {}).length > 1}
        <FieldConfigPopover
          {fieldConfig}
          label={config.label ?? key}
          onChange={updateFieldProperty}
          {chartFieldInput}
        />
      {/if}
    {/key}
  </div>

  <SingleFieldInput
    {canvasName}
    metricName={metricsView}
    id={`${key}-field`}
    type={isDimension ? "dimension" : "measure"}
    includeTime={!chartFieldInput?.hideTimeDimension}
    selectedItem={fieldConfig?.field}
    onSelect={async (field) => {
      updateFieldConfig(field);
    }}
  />
</div>
