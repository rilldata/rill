<!-- @component
This component maps the input params for a component to a form input.
It is used in the ComponentsEditor component to render the input fields for the selected component. 
-->

<script lang="ts">
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import { getComponentObj } from "@rilldata/web-common/features/canvas/components/util";
  import MetricSelectorDropdown from "@rilldata/web-common/features/canvas/inspector/MetricSelectorDropdown.svelte";
  import { getCanvasStateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import type { CanvasComponentType } from "../components/types";
  import FieldSelectorDropdown from "./FieldSelectorDropdown.svelte";
  import PositionalFieldConfig from "./PositionalFieldConfig.svelte";

  export let componentType: CanvasComponentType;
  export let paramValues: Record<string, any>;

  const { fileArtifact, canvasStore } = getCanvasStateManagers();

  $: selectedComponentIndex = $canvasStore?.selectedComponentIndex || 0;

  $: path = ["items", selectedComponentIndex, "component", componentType];

  $: component = getComponentObj(
    $fileArtifact,
    path,
    componentType,
    paramValues,
  );

  $: inputParams = component.inputParams();

  $: spec = component.specStore;
  $: metricsView = "metrics_view" in $spec ? $spec.metrics_view : null;
</script>

<div>
  {#each Object.entries(inputParams) as [key, config]}
    {#if config.showInUI !== false}
      {#if config.type === "text" || config.type === "number" || config.type === "rill_time"}
        <Input
          inputType={config.type === "number" ? "number" : "text"}
          capitalizeLabel={false}
          textClass="text-sm"
          optional={config.required == false}
          label={config.label || key}
          bind:value={paramValues[key]}
          onBlur={async () => {
            component.updateProperty(key, paramValues[key]);
          }}
          onEnter={async () => {
            component.updateProperty(key, paramValues[key]);
          }}
        />
      {:else if config.type === "metrics"}
        <MetricSelectorDropdown {component} {key} inputParam={config} />
      {:else if metricsView && (config.type === "measure" || config.type === "dimension")}
        <FieldSelectorDropdown
          label={config.label || key}
          metricName={metricsView}
          id={key}
          type={config.type}
          selectedItem={paramValues[key]}
          onSelect={async (field) => {
            component.updateProperty(key, field);
          }}
        />
      {:else if config.type === "textArea"}
        <textarea
          class="w-full p-2 border border-gray-300 rounded-sm"
          rows="4"
          bind:value={paramValues[key]}
          on:blur={async () => {
            component.updateProperty(key, paramValues[key]);
          }}
          placeholder={config.label || key}
        ></textarea>
      {:else if metricsView && config.type === "positional"}
        <PositionalFieldConfig
          {key}
          {config}
          {metricsView}
          value={paramValues[key] || {}}
          onChange={(updatedConfig) => {
            paramValues[key] = updatedConfig;
            component.updateProperty(key, updatedConfig);
          }}
        />
      {/if}
    {/if}
  {/each}
</div>
