<script lang="ts">
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import ChartTypeSelector from "@rilldata/web-common/features/canvas/inspector/ChartTypeSelector.svelte";
  import MetricSelectorDropdown from "@rilldata/web-common/features/canvas/inspector/MetricSelectorDropdown.svelte";
  import FieldSelectorDropdown from "./FieldSelectorDropdown.svelte";
  import PositionalFieldConfig from "./PositionalFieldConfig.svelte";

  import { getComponentObj } from "@rilldata/web-common/features/canvas/components/util";
  import { getCanvasStateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";

  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import { type V1ComponentSpecRendererProperties } from "@rilldata/web-common/runtime-client";
  import { onMount } from "svelte";
  import type { CanvasComponentType } from "../components/types";

  export let componentType: CanvasComponentType;
  export let paramValues: V1ComponentSpecRendererProperties;
  export let fileArtifact: FileArtifact;

  const { canvasStore } = getCanvasStateManagers();
  $: selectedComponentIndex = $canvasStore?.selectedComponentIndex ?? 0;
  $: path = ["items", selectedComponentIndex, "component", componentType];

  $: localParamValues = localParamValues || {};
  let oldParamValuesRef: V1ComponentSpecRendererProperties = {};

  // TODO: Make this robust possibly a store.
  $: if (JSON.stringify(paramValues) !== JSON.stringify(oldParamValuesRef)) {
    localParamValues = structuredClone(paramValues) || {};
    oldParamValuesRef = paramValues;
  }

  onMount(() => {
    localParamValues = structuredClone(paramValues) || {};
  });

  $: component = getComponentObj(
    fileArtifact,
    path,
    componentType,
    paramValues,
  );
  $: inputParams = component.inputParams();
  $: spec = component.specStore;

  $: metricsView = "metrics_view" in $spec ? $spec.metrics_view : null;
</script>

<ChartTypeSelector {component} {componentType} />

{#key selectedComponentIndex}
  <div>
    {#each Object.entries(inputParams) as [key, config]}
      {#if config.showInUI !== false}
        <!-- TEXT, NUMBER, RILL_TIME -->
        {#if config.type === "text" || config.type === "number" || config.type === "rill_time"}
          <Input
            inputType={config.type === "number" ? "number" : "text"}
            capitalizeLabel={false}
            textClass="text-sm"
            optional={!config.required}
            label={config.label ?? key}
            bind:value={localParamValues[key]}
            onBlur={async () => {
              component.updateProperty(key, localParamValues[key]);
            }}
            onEnter={async () => {
              component.updateProperty(key, localParamValues[key]);
            }}
          />

          <!-- METRICS SELECTOR -->
        {:else if config.type === "metrics"}
          <MetricSelectorDropdown {component} {key} inputParam={config} />

          <!-- MEASURE / DIMENSION with metricsView -->
        {:else if metricsView && (config.type === "measure" || config.type === "dimension")}
          <FieldSelectorDropdown
            label={config.label ?? key}
            metricName={metricsView}
            id={key}
            type={config.type}
            selectedItem={localParamValues[key]}
            onSelect={async (field) => {
              component.updateProperty(key, field);
            }}
          />

          <!-- BOOLEAN SWITCH -->
        {:else if config.type === "boolean"}
          <div class="flex items-center justify-between py-2">
            <InputLabel
              label={config.label ?? key}
              optional={!config.required}
              id={key}
            />
            <Switch
              bind:checked={localParamValues[key]}
              on:click={async () => {
                component.updateProperty(key, localParamValues[key]);
              }}
              small
            />
          </div>

          <!-- TEXT AREA -->
        {:else if config.type === "textArea"}
          <InputLabel
            label={config.label ?? key}
            optional={!config.required}
            id={key}
          />
          <textarea
            class="w-full p-2 border border-gray-300 rounded-sm"
            rows="4"
            bind:value={localParamValues[key]}
            on:blur={async () => {
              component.updateProperty(key, localParamValues[key]);
            }}
            placeholder={config.label ?? key}
          />

          <!-- POSITIONAL CONFIG -->
        {:else if metricsView && config.type === "positional"}
          <PositionalFieldConfig
            {key}
            {config}
            {metricsView}
            value={localParamValues[key] || {}}
            onChange={(updatedConfig) => {
              localParamValues[key] = updatedConfig;
              component.updateProperty(key, updatedConfig);
            }}
          />
        {/if}
      {/if}
    {/each}
  </div>
{/key}
