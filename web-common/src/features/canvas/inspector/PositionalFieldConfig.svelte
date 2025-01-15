<script lang="ts">
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import SingleFieldInput from "@rilldata/web-common/features/canvas/inspector/SingleFieldInput.svelte";
  import type { FieldConfig } from "../components/charts/types";

  export let key: string;
  export let config: { label?: string };
  export let metricsView: string;
  export let fieldConfig: FieldConfig;
  export let onChange: (updatedConfig: FieldConfig) => void;

  $: isDimension = key === "x";

  function updateFieldConfig(fieldName: string) {
    const updatedConfig: FieldConfig = {
      ...fieldConfig,
      field: fieldName,
      type: isDimension ? "nominal" : "quantitative",
    };

    // TODO: Add displayName to title
    onChange(updatedConfig);
  }

  function updateFieldProperty(property: keyof FieldConfig, value: any) {
    const updatedConfig: FieldConfig = {
      ...fieldConfig,
      [property]: value,
    };

    onChange(updatedConfig);
  }

  let isDropdownOpen = false;
</script>

<div class="gap-y-1">
  <div class="flex justify-between items-center">
    <InputLabel small label={config.label ?? key} id={key} />
    <DropdownMenu.Root bind:open={isDropdownOpen}>
      <DropdownMenu.Trigger class="flex-none">
        <IconButton rounded active={isDropdownOpen}>
          <ThreeDot size="16px" />
        </IconButton>
      </DropdownMenu.Trigger>
      <DropdownMenu.Content align="start" class="w-[250px]">
        <DropdownMenu.CheckboxItem
          checked={fieldConfig?.showAxisTitle}
          on:click={async () => {
            updateFieldProperty("showAxisTitle", !fieldConfig?.showAxisTitle);
          }}
        >
          <span class="ml-2">Show axis title</span>
        </DropdownMenu.CheckboxItem>
      </DropdownMenu.Content>
    </DropdownMenu.Root>
  </div>

  <SingleFieldInput
    metricName={metricsView}
    id={`${key}-field`}
    type={isDimension ? "dimension" : "measure"}
    selectedItem={fieldConfig?.field}
    onSelect={async (field) => {
      updateFieldConfig(field);
    }}
  />
</div>
