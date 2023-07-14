<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import type { V1Model } from "@rilldata/web-common/runtime-client";
  import type { Readable } from "svelte/store";
  import {
    addQuickMetricsToDashboardYAML,
    MetricsInternalRepresentation,
  } from "../../metrics-internal-store";
  import QuickMetricsModal from "../QuickMetricsModal.svelte";

  export let selectedModel: V1Model;
  export let metricsInternalRep: Readable<MetricsInternalRepresentation>;
  export let handlePutAndMigrate;

  $: measures = $metricsInternalRep.getMeasures();
  $: dimensions = $metricsInternalRep.getDimensions();

  async function handleGenerateClick(yaml: string) {
    // if the timeseries field is empty or does not exist,
    // add in the first timestamp column available.
    const newYAMLString = addQuickMetricsToDashboardYAML(yaml, selectedModel);
    handlePutAndMigrate(newYAMLString);

    // invalidateMetricsView(queryClient, metricsDefId);
    // // In `svelte-query/totals.ts`, in the `invalidateMetricsViewData()` function, we use `refetchQueries` where we should probably use `invalidateQueries`.
    // // We should make that change, but it has a wide surface area, so we need to take the time to QA it properly. In the meantime, we need to remove old
    // // queryKeys from the queryCache, so they don't get refetched and generate 500 errors.
    // queryClient.removeQueries([`v1/metrics-view/toplist`, metricsDefId], {
    //   exact: false,
    // });

    closeModal();
  }

  let tooltipText = "";
  let buttonDisabled = true;
  $: if ($metricsInternalRep.getMetricKey("model") === "") {
    tooltipText = "Select a model before populating these metrics";
    buttonDisabled = true;
  } else {
    tooltipText = undefined;
    buttonDisabled = false;
  }

  let modalIsOpen = false;

  const openModelIfNeeded = () => {
    if (measures?.length > 0 || dimensions?.length > 0) {
      openModal();
    } else {
      handleGenerateClick($metricsInternalRep.internalYAML);
    }
  };

  const openModal = () => {
    modalIsOpen = true;
  };

  const closeModal = () => {
    modalIsOpen = false;
  };
</script>

<Tooltip alignment="middle" distance={5} location="right">
  <button
    class={`bg-white
          border-gray-400
          hover:border-gray-900
          transition-tranform
          duration-100
          items-center
          justify-center
          border
          rounded
          flex flex-row gap-x-2
          pl-4 pr-4
          pt-2 pb-2
          ${buttonDisabled ? "cursor-not-allowed" : "cursor-pointer"}
          ${buttonDisabled ? "text-gray-500" : "text-gray-900"}
        `}
    disabled={buttonDisabled}
    on:click={openModelIfNeeded}>Quick Metrics</button
  >
  <TooltipContent slot="tooltip-content">
    <div>
      {#if buttonDisabled}
        {tooltipText}
      {:else}
        <div style="max-width: 30em;">
          Add initial measure <em>events per time period</em>, and add all
          categorical columns as slicing dimensions. If timestamp is present,
          the first time column from the model will be selected.
          <br /> <strong>Warning:</strong> Replaces current measures and dimensions.
        </div>
      {/if}
    </div>
  </TooltipContent>
</Tooltip>

{#if modalIsOpen}
  <QuickMetricsModal
    on:cancel={closeModal}
    on:click-outside={closeModal}
    on:replace-metrics={() =>
      handleGenerateClick($metricsInternalRep.internalYAML)}
  />
{/if}
