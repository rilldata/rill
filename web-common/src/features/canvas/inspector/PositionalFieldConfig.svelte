<script lang="ts">
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
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
  <Input
    inputType="text"
    capitalizeLabel={false}
    textClass="text-sm"
    label={`${config.label || key} Label`}
    bind:value={value.label}
    onBlur={async () => {
      updateFieldConfig("label", value.label);
    }}
  />
  <Input
    inputType="text"
    capitalizeLabel={false}
    textClass="text-sm"
    label={`${config.label || key} Format`}
    bind:value={value.format}
    onBlur={async () => {
      updateFieldConfig("format", value.format);
    }}
  />
</div>
