<script lang="ts">
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import { BaseChart } from "@rilldata/web-common/features/canvas/components/charts/BaseChart";
  import type { BaseCanvasComponent } from "../components/BaseCanvasComponent";
  import { PivotCanvasComponent } from "../components/pivot";
  import type { ComponentSpec } from "../components/types";
  import AlignmentInput from "./AlignmentInput.svelte";
  import CanvasFieldSwitcher from "./CanvasFieldSwitcher.svelte";
  import ChartTypeSelector from "./chart/ChartTypeSelector.svelte";
  import MarkSelector from "./chart/MarkSelector.svelte";
  import PositionalFieldConfig from "./chart/PositionalFieldConfig.svelte";
  import ComparisonInput from "./ComparisonInput.svelte";
  import MultiFieldInput from "./fields/MultiFieldInput.svelte";
  import SingleFieldInput from "./fields/SingleFieldInput.svelte";
  import MetricSelectorDropdown from "./MetricSelectorDropdown.svelte";
  import SparklineInput from "./SparklineInput.svelte";
  import TableTypeSelector from "./TableTypeSelector.svelte";
  import type { AllKeys, ComponentInputParam } from "./types";

  export let component: BaseCanvasComponent;

  $: ({
    specStore,
    parent: { name: canvasName },
  } = component);

  $: localParamValues = $specStore;

  $: inputParams = component.inputParams(
    component instanceof PivotCanvasComponent
      ? "columns" in $specStore
        ? "table"
        : "pivot"
      : undefined,
  ).options;

  $: metricsView =
    "metrics_view" in localParamValues ? localParamValues.metrics_view : null;

  $: entries = Object.entries(inputParams) as [
    AllKeys<ComponentSpec>,
    ComponentInputParam,
  ][];
</script>

{#if component instanceof BaseChart}
  <ChartTypeSelector {component} />
{/if}

{#if metricsView && component instanceof PivotCanvasComponent}
  <TableTypeSelector {component} />
{/if}

<div>
  {#each entries as [key, config] (`${component.id}-${key}`)}
    {#if config.showInUI !== false}
      <div
        class="component-param"
        class:grouped={config.meta?.layout === "grouped"}
      >
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
            bind:value={$specStore[key]}
            onBlur={() => {
              component.updateProperty(key, localParamValues[key]);
            }}
            onEnter={() => {
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
            onSelect={(field) => {
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
            onMultiSelect={(field) => {
              component.updateProperty(key, field);
            }}
          />

          <!-- BOOLEAN SWITCH -->
        {:else if config.type === "boolean"}
          <div class="flex items-center justify-between py-1">
            <InputLabel
              small
              label={config.label ?? key}
              id={key}
              faint={config.meta?.invertBoolean
                ? localParamValues[key]
                : !localParamValues[key]}
            />
            <Switch
              checked={config.meta?.invertBoolean
                ? !$specStore[key]
                : $specStore[key]}
              on:click={() => {
                component.updateProperty(key, !localParamValues[key]);
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
              bind:value={$specStore[key]}
              on:blur={() => {
                component.updateProperty(key, localParamValues[key]);
              }}
              placeholder={config.label ?? key}
            />
          </div>

          <!-- SELECT DROPDOWN -->
        {:else if config.type === "select"}
          <Select
            id={key}
            label={config.label ?? key}
            options={config.meta?.options ?? []}
            value={$specStore[key] ?? config.meta?.default}
            full={true}
            size="sm"
            sameWidth
            fontSize={12}
            onChange={(newValue) => {
              component.updateProperty(key, newValue);
            }}
          />

          <!-- SWITCHER TABS -->
        {:else if config.type === "switcher_tab"}
          <CanvasFieldSwitcher
            {key}
            label={config.label ?? key}
            options={config.meta?.options ?? []}
            value={localParamValues[key] ?? config.meta?.default}
            onChange={(newValue) => {
              component.updateProperty(key, newValue);
            }}
          />

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
            {key}
            {config}
            {metricsView}
            markConfig={localParamValues[key] || {}}
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
    @apply border-t;
  }
  .component-param.grouped {
    @apply py-0;
    @apply border-none;
  }
</style>
