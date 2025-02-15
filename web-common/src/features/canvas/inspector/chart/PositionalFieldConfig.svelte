<script lang="ts">
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import type { FieldConfig } from "@rilldata/web-common/features/canvas/components/charts/types";
  import SingleFieldInput from "@rilldata/web-common/features/canvas/inspector/SingleFieldInput.svelte";
  import { getCanvasStateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";

  export let key: string;
  export let config: { label?: string };
  export let metricsView: string;
  export let fieldConfig: FieldConfig;
  export let onChange: (updatedConfig: FieldConfig) => void;

  const {
    canvasEntity: {
      spec: { getTimeDimensionForMetricView },
    },
  } = getCanvasStateManagers();

  $: isDimension = key === "x";
  $: timeDimension = getTimeDimensionForMetricView(metricsView);

  function updateFieldConfig(fieldName: string) {
    const isTime = $timeDimension && fieldName === $timeDimension;

    let updatedConfig: FieldConfig;
    if (isTime && $timeDimension) {
      updatedConfig = {
        ...fieldConfig,
        field: $timeDimension,
        type: "temporal",
      };
    } else {
      updatedConfig = {
        ...fieldConfig,
        field: fieldName,
        type: isTime ? "temporal" : isDimension ? "nominal" : "quantitative",
      };
    }

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
        <div class="px-2 py-1.5 flex items-center justify-between">
          <span class="text-xs">Show axis title</span>
          <Switch
            small
            checked={fieldConfig?.showAxisTitle}
            on:click={() => {
              updateFieldProperty("showAxisTitle", !fieldConfig?.showAxisTitle);
            }}
          />
        </div>
        {#if !isDimension}
          <div class="px-2 py-1.5 flex items-center justify-between">
            <span class="text-xs">Zero based origin</span>
            <Switch
              small
              checked={fieldConfig?.zeroBasedOrigin}
              on:click={() => {
                updateFieldProperty(
                  "zeroBasedOrigin",
                  !fieldConfig?.zeroBasedOrigin,
                );
              }}
            />
          </div>
        {/if}
      </DropdownMenu.Content>
    </DropdownMenu.Root>
  </div>

  <SingleFieldInput
    metricName={metricsView}
    id={`${key}-field`}
    type={isDimension ? "dimension" : "measure"}
    includeTime
    selectedItem={fieldConfig?.field}
    onSelect={async (field) => {
      updateFieldConfig(field);
    }}
  />
</div>
