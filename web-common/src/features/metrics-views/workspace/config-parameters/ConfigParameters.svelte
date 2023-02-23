<script lang="ts">
  import Callout from "@rilldata/web-common/components/callout/Callout.svelte";
  import type { V1Model } from "@rilldata/web-common/runtime-client";
  import type { Readable } from "svelte/store";
  import type { MetricsInternalRepresentation } from "../../metrics-internal-store";
  import DefaultTimeRangeSelector from "./DefaultTimeRangeSelector.svelte";
  import DisplayNameInput from "./DisplayNameInput.svelte";
  import ModelSelector from "./ModelSelector.svelte";
  import QuickStartButton from "./QuickStartButton.svelte";
  import SmallestTimeGrainSelector from "./SmallestTimeGrainSelector.svelte";
  import TimeColumnSelector from "./TimeColumnSelector.svelte";
  export let workspaceWidth: number;

  export let metricsSourceSelectionError;
  export let metricsInternalRep: Readable<MetricsInternalRepresentation>;
  export let model: V1Model;
  export let updateRuntime: () => void;

  let gridTemplate = "repeat(3, 45px)";
  $: metricsConfigWidth = workspaceWidth || 0;
  $: gridTemplate =
    metricsConfigWidth < 1400 ? "repeat(3, 35px)" : "repeat(2, 40px)";

  $: timeColumn = $metricsInternalRep.getMetricKey("timeseries");
</script>

<div class="flex-none flex flex-row">
  <div
    style:grid-template-rows={gridTemplate}
    class="grid grid-flow-col gap-y-2 gap-x-8"
  >
    <DisplayNameInput {metricsInternalRep} />
    <ModelSelector {metricsInternalRep} />
    <TimeColumnSelector selectedModel={model} {metricsInternalRep} />
    {#if timeColumn}
      <SmallestTimeGrainSelector selectedModel={model} {metricsInternalRep} />
      <DefaultTimeRangeSelector selectedModel={model} {metricsInternalRep} />
    {/if}
  </div>
  <div class="ml-auto">
    {#if metricsSourceSelectionError}
      <Callout level="error">
        {metricsSourceSelectionError}
      </Callout>
    {:else}
      <QuickStartButton
        handlePutAndMigrate={updateRuntime}
        selectedModel={model}
        {metricsInternalRep}
      />
    {/if}
  </div>
</div>
