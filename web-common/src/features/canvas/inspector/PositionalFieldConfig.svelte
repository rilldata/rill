<script lang="ts">
  import type { FieldConfig } from "../components/charts/types";
  import FieldSelectorDropdown from "./FieldSelectorDropdown.svelte";

  export let key: string;
  export let config: { label?: string };
  export let metricsView: string;
  export let value: FieldConfig;
  export let onChange: (updatedConfig: FieldConfig) => void;

  function updateFieldConfig(property: keyof FieldConfig, field: string) {
    const updatedConfig: FieldConfig = {
      ...value,
      [property]: field,
    };
    if (!updatedConfig.type) {
      updatedConfig.type = "quantitative"; // Default type for measures
    }
    onChange(updatedConfig);
  }
</script>

<div class="space-y-2">
  <FieldSelectorDropdown
    label={`${config.label || key} Field`}
    metricName={metricsView}
    id={`${key}-field`}
    type="measure"
    selectedItem={value?.field || ""}
    onSelect={async (field) => {
      updateFieldConfig("field", field);
    }}
  />
</div>
