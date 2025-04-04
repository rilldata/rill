<script lang="ts">
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import {
    isChartComponentType,
    isTableComponentType,
    type CanvasComponentObj,
  } from "@rilldata/web-common/features/canvas/components/util";
  import { type V1ComponentSpecRendererProperties } from "@rilldata/web-common/runtime-client";
  import { onMount } from "svelte";
  import type { CanvasComponentType } from "../components/types";
  import AlignmentInput from "./AlignmentInput.svelte";
  import ChartTypeSelector from "./chart/ChartTypeSelector.svelte";
  import MarkSelector from "./chart/MarkSelector.svelte";
  import PositionalFieldConfig from "./chart/PositionalFieldConfig.svelte";
  import ComparisonInput from "./ComparisonInput.svelte";
  import MetricSelectorDropdown from "./MetricSelectorDropdown.svelte";
  import MultiFieldInput from "./MultiFieldInput.svelte";
  import SingleFieldInput from "./SingleFieldInput.svelte";
  import SparklineInput from "./SparklineInput.svelte";
  import TableTypeSelector from "./TableTypeSelector.svelte";

  export let component: CanvasComponentObj;
  export let componentType: CanvasComponentType;
  export let paramValues: V1ComponentSpecRendererProperties;
  export let canvasName: string;

  $: localParamValues = localParamValues || {};
  let oldParamValuesRef: V1ComponentSpecRendererProperties = {};

  // TODO: Make this robust possibly a store.
  $: if (JSON.stringify(paramValues) !== JSON.stringify(oldParamValuesRef)) {
    localParamValues = structuredClone(paramValues) || {};
    oldParamValuesRef = paramValues;
  }

  $: inputParams = component.inputParams().options;

  $: metricsView =
    "metrics_view" in paramValues ? (paramValues.metrics_view as string) : null;

  onMount(() => {
    localParamValues = structuredClone(paramValues) || {};
  });
</script>

{#if isChartComponentType(componentType)}
  <ChartTypeSelector {component} {componentType} />
{/if}

{#if metricsView && isTableComponentType(componentType)}
  <TableTypeSelector
    {canvasName}
    {component}
    {componentType}
    metricsViewName={metricsView}
  />
{/if}

<div>
  {#each Object.entries(inputParams) as [key, config] (key)}
    {#if config.showInUI !== false}
      <div class="component-param">
        <!-- TEXT, NUMBER, RILL_TIME -->
        {#if config.type === "text" || config.type === "number" || config.type === "rill_time"}
          <Input
            inputType={config.type === "number" ? "number" : "text"}
            capitalizeLabel={false}
            textClass="text-sm"
            size="sm"
            placeholder={config?.meta?.placeholder ?? ""}
            labelGap={2}
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

          <!-- MEASURE / DIMENSION -->
        {:else if metricsView && (config.type === "measure" || config.type === "dimension")}
          <SingleFieldInput
            {canvasName}
            label={config.label ?? key}
            metricName={metricsView}
            id={key}
            type={config.type}
            selectedItem={localParamValues[key]}
            onSelect={async (field) => {
              component.updateProperty(key, field);
            }}
          />

          <!-- MULTIPLE MEASURE / MULTIPLE DIMENSION / MULTIPLE FIELDS -->
        {:else if metricsView && config.type === "multi_fields"}
          <MultiFieldInput
            {canvasName}
            label={config.label ?? key}
            metricName={metricsView}
            id={key}
            types={config.meta?.allowedTypes ?? ["measure", "dimension"]}
            selectedItems={localParamValues[key]}
            onMultiSelect={async (field) => {
              component.updateProperty(key, field);
            }}
          />

          <!-- BOOLEAN SWITCH -->
        {:else if config.type === "boolean"}
          <div class="flex items-center justify-between py-2">
            <InputLabel
              small
              label={config.label ?? key}
              id={key}
              faint={!localParamValues[key]}
            />
            <Switch
              bind:checked={localParamValues[key]}
              on:click={async () => {
                component.updateProperty(key, !paramValues[key]);
              }}
              small
            />
          </div>

          <!-- TEXT AREA -->
        {:else if config.type === "textArea"}
          <div class="flex flex-col gap-y-2">
            <InputLabel
              hint={config?.description}
              small
              label={config.label ?? key}
              id={key}
            />
            <textarea
              class="w-full p-2 border border-gray-300 rounded-sm"
              rows="8"
              bind:value={localParamValues[key]}
              on:blur={async () => {
                component.updateProperty(key, localParamValues[key]);
              }}
              placeholder={config.label ?? key}
            />
          </div>

          <!-- KPI SPARKLINE INPUT -->
        {:else if config.type === "sparkline"}
          <SparklineInput
            {key}
            label={config.label ?? key}
            value={localParamValues[key]}
            onChange={(updatedSparkline) => {
              localParamValues[key] = updatedSparkline;
              component.updateProperty(key, updatedSparkline);
            }}
          />

          <!-- COMPARISON OPTIONS INPUT -->
        {:else if config.type === "comparison_options"}
          <ComparisonInput
            {key}
            label={config.label ?? key}
            options={localParamValues[key]}
            onChange={(options) => {
              localParamValues[key] = options;
              component.updateProperty(key, options);
            }}
          />

          <!-- COMPONENT CONTENTS ALIGNMENT -->
        {:else if config.type === "alignment"}
          <AlignmentInput
            {key}
            label={config.label ?? key}
            position={localParamValues[key]}
            defaultAlignment={config.meta?.defaultAlignment}
            onChange={(updatedPosition) => {
              localParamValues[key] = updatedPosition;
              component.updateProperty(key, updatedPosition);
            }}
          />
          <!-- POSITIONAL CONFIG -->
        {:else if metricsView && config.type === "positional"}
          <PositionalFieldConfig
            {canvasName}
            {key}
            {config}
            {metricsView}
            fieldConfig={localParamValues[key] || {}}
            onChange={(updatedConfig) => {
              localParamValues[key] = updatedConfig;
              component.updateProperty(key, updatedConfig);
            }}
          />
          <!-- COLOR CONFIG -->
        {:else if metricsView && config.type === "mark"}
          <MarkSelector
            {canvasName}
            label={config.label ?? key}
            {key}
            {metricsView}
            value={localParamValues[key] || {}}
            onChange={(updatedConfig) => {
              localParamValues[key] = updatedConfig;
              component.updateProperty(key, updatedConfig);
            }}
          />
        {/if}
      </div>
    {/if}
  {/each}
</div>

<style lang="postcss">
  .component-param {
    @apply py-3 px-5;
    @apply border-t border-gray-200;
  }
</style>
