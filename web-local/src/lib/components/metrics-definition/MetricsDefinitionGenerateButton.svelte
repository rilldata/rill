<script lang="ts">
  import type { V1Model } from "@rilldata/web-common/runtime-client";
  import type { Readable } from "svelte/store";
  import {
    generateMeasuresAndDimension,
    MetricsInternalRepresentation,
  } from "../../application-state-stores/metrics-internal-store";
  import { selectTimestampColumnFromSchema } from "../../svelte-query/column-selectors";
  import Tooltip from "../tooltip/Tooltip.svelte";
  import TooltipContent from "../tooltip/TooltipContent.svelte";
  import QuickMetricsModal from "./QuickMetricsModal.svelte";

  $: measures = $metricsInternalRep.getMeasures();
  $: dimensions = $metricsInternalRep.getDimensions();

  export let selectedModel: V1Model;
  export let metricsInternalRep: Readable<MetricsInternalRepresentation>;
  export let handlePutAndMigrate;

  async function handleGenerateClick() {
    // if the timeseries field is empty or does not exist,
    // add in the first timestamp column available.
    // if no timestamp column available, we currently do nothing in this case.
    // later, we'll remove the requiremen t for a timeseries field.
    const newYAMLString = generateMeasuresAndDimension(selectedModel, {
      timeseries:
        $metricsInternalRep.getMetricKey("timeseries") || timestampColumns[0],
    });
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

  let timestampColumns: Array<string>;
  $: if (selectedModel) {
    timestampColumns = selectTimestampColumnFromSchema(selectedModel?.schema);
  } else {
    timestampColumns = [];
  }

  let tooltipText = "";
  let buttonDisabled = true;
  $: if ($metricsInternalRep.getMetricKey("model") === "") {
    tooltipText = "Select a model before populating these metrics";
    buttonDisabled = true;
  } else if (timestampColumns.length === 0) {
    tooltipText = "Cannot create metrics for a model with no timestamps";
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
      handleGenerateClick();
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
    on:click={openModelIfNeeded}>quick metrics</button
  >
  <TooltipContent slot="tooltip-content">
    <div>
      {#if buttonDisabled}
        {tooltipText}
      {:else}
        <div style="max-width: 30em;">
          Add initial measure <em>events per time period</em>, and add all
          categorical columns as slicing dimensions. If no timestamp is
          selected, the first time column from the model will be used.
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
    on:replace-metrics={handleGenerateClick}
  />
{/if}
