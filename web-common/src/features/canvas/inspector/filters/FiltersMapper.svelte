<script lang="ts">
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import { type CanvasComponentObj } from "@rilldata/web-common/features/canvas/components/util";
  import DimensionFiltersInput from "@rilldata/web-common/features/canvas/inspector/filters/DimensionFiltersInput.svelte";
  import { type V1ComponentSpecRendererProperties } from "@rilldata/web-common/runtime-client";
  import { onMount } from "svelte";

  export let selectedComponentName: string;
  export let component: CanvasComponentObj;
  export let paramValues: V1ComponentSpecRendererProperties;

  $: localParamValues = localParamValues || {};
  let oldParamValuesRef: V1ComponentSpecRendererProperties = {};

  // TODO: Make this robust possibly a store.
  $: if (JSON.stringify(paramValues) !== JSON.stringify(oldParamValuesRef)) {
    localParamValues = structuredClone(paramValues) || {};
    oldParamValuesRef = paramValues;
  }

  $: inputParams = component.inputParams().filter;

  // Enable this when adding support for dimension filters
  $: metricsView =
    "metrics_view" in paramValues ? paramValues.metrics_view : null;

  onMount(() => {
    localParamValues = structuredClone(paramValues) || {};
  });
</script>

<div>
  {#each Object.entries(inputParams) as [key, config]}
    <div class="component-param">
      {#if config.type === "time_range" || config.type === "comparison_range"}
        <Input
          inputType="text"
          capitalizeLabel={false}
          textClass="text-sm"
          size="sm"
          labelGap={2}
          optional
          label={config.label ?? key}
          bind:value={localParamValues[key]}
          onBlur={async () => {
            component.updateProperty(key, localParamValues[key]);
          }}
          onEnter={async () => {
            component.updateProperty(key, localParamValues[key]);
          }}
        />
      {:else if config.type == "dimension_filters" && metricsView}
        <DimensionFiltersInput
          {metricsView}
          {selectedComponentName}
          label={config.label ?? key}
          id={key}
          filter={localParamValues[key]}
          onChange={(filter) => {
            console.log("before", localParamValues[key], "filter", filter);
            localParamValues[key] = filter;
            component.updateProperty(key, filter);
          }}
        />
      {/if}
    </div>
  {/each}
</div>

<style lang="postcss">
  .component-param {
    @apply py-3 px-5;
    @apply border-t border-gray-200;
  }
</style>
