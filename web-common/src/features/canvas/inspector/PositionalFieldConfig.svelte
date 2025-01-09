<script lang="ts">
  import SingleFieldInput from "@rilldata/web-common/features/canvas/inspector/SingleFieldInput.svelte";
  import type { FieldConfig } from "../components/charts/types";

  export let key: string;
  export let config: { label?: string };
  export let metricsView: string;
  export let value: FieldConfig;
  export let onChange: (updatedConfig: FieldConfig) => void;

  $: isDimension = key === "x";

  function updateFieldConfig(field: string) {
    const updatedConfig: FieldConfig = {
      ...value,
      field,
      type: isDimension ? "nominal" : "quantitative",
    };
    onChange(updatedConfig);
  }
</script>

<div class="space-y-2">
  <SingleFieldInput
    label={config.label || key}
    metricName={metricsView}
    id={`${key}-field`}
    type={isDimension ? "dimension" : "measure"}
    selectedItem={value?.field}
    onSelect={async (field) => {
      updateFieldConfig(field);
    }}
  />
</div>
