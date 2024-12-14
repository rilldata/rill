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

  export let componentType: CanvasComponentType;
  export let params: Record<string, any>;

  const { fileArtifact, canvasStore } = getCanvasStateManagers();

  $: selectedComponentIndex = $canvasStore?.selectedComponentIndex || 0;

  $: path = ["items", selectedComponentIndex, "component", componentType];

  $: component = getComponentObj($fileArtifact, path, componentType, params);
  $: inputParams = component.inputParams();

  $: spec = component.specStore;

  $: metricsView = "metrics_view" in $spec ? $spec.metrics_view : null;
</script>

<div>
  {#each Object.entries(inputParams) as [key, params]}
    {#if params.showInUI !== false}
      {#if params.type === "text" || params.type === "number" || params.type === "rill_time"}
        <Input
          inputType={params.type === "number" ? "number" : "text"}
          capitalizeLabel={false}
          textClass="text-sm"
          optional={params.required == false}
          label={params.label || key}
          bind:value={params[key]}
          onBlur={async () => {
            component.updateProperty(key, params[key]);
          }}
          onEnter={async () => {
            component.updateProperty(key, params[key]);
          }}
        />
      {:else if params.type === "metrics_view"}
        <MetricSelectorDropdown {component} {key} {params} />
      {:else if metricsView && (params.type === "measure" || params.type === "dimension")}
        <FieldSelectorDropdown
          label={params.label || key}
          metricName={metricsView}
          id={key}
          type={params.type}
          selectedItem={params[key]}
          onSelect={async (field) => {
            component.updateProperty(key, field);
          }}
        />
      {:else if params.type === "textArea"}
        <textarea
          class="w-full p-2 border border-gray-300 rounded-sm"
          rows="4"
          value={params[key]}
          on:blur={async () => {
            component.updateProperty(key, params[key]);
          }}
          placeholder={params.label || key}
        ></textarea>
      {/if}
    {/if}
  {/each}
</div>
