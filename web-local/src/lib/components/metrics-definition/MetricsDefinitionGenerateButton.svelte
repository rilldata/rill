<script lang="ts">
  import { getContext } from "svelte";
  import type { Readable } from "svelte/store";
  import { useQueryClient } from "@sveltestack/svelte-query";
  import {
    useRuntimeServicePutFileAndMigrate,
    V1Model,
  } from "@rilldata/web-common/runtime-client";
  import type { DerivedModelStore } from "../../application-state-stores/model-stores";
  import { TIMESTAMPS } from "../../duckdb-data-types";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { invalidateMetricsView } from "../../svelte-query/queries/metrics-views/invalidation";
  import Tooltip from "../tooltip/Tooltip.svelte";
  import TooltipContent from "../tooltip/TooltipContent.svelte";
  import {
    generateMeasuresAndDimension,
    MetricsInternalRepresentation,
  } from "../../application-state-stores/metrics-internal-store";
  import QuickMetricsModal from "./QuickMetricsModal.svelte";

  $: measures = $metricsInternalRep.getMeasures();
  $: dimensions = $metricsInternalRep.getDimensions();

  export let selectedModel: V1Model;
  export let metricsInternalRep: Readable<MetricsInternalRepresentation>;
  export let handlePutAndMigrate;

  const queryClient = useQueryClient();

  async function handleGenerateClick() {
    // await store.dispatch(generateMeasuresAndDimensionsApi(metricsDefId));

    const newYAMLString = generateMeasuresAndDimension(selectedModel);

    handlePutAndMigrate(newYAMLString);

    if (
      !$metricsInternalRep.getMetricKey("timeseries") &&
      timestampColumns.length > 0
    ) {
      // select the first available timestamp column if one has not been
      // selected and there are some available
      $metricsInternalRep.updateMetricKey("timeseries", timestampColumns[0]);
    }

    // invalidateMetricsView(queryClient, metricsDefId);
    // // In `svelte-query/totals.ts`, in the `invalidateMetricsViewData()` function, we use `refetchQueries` where we should probably use `invalidateQueries`.
    // // We should make that change, but it has a wide surface area, so we need to take the time to QA it properly. In the meantime, we need to remove old
    // // queryKeys from the queryCache, so they don't get refetched and generate 500 errors.
    // queryClient.removeQueries([`v1/metrics-view/toplist`, metricsDefId], {
    //   exact: false,
    // });

    closeModal();
  }

  const derivedModelStore = getContext(
    "rill:app:derived-model-store"
  ) as DerivedModelStore;

  let timestampColumns: Array<string>;
  $: if (selectedModel) {
    const selectedMetricsDefModelProfile = selectedModel?.schema?.fields ?? [];
    timestampColumns = selectedMetricsDefModelProfile
      .filter((column) => TIMESTAMPS.has(column.type.code as string))
      .map((column) => column.name);
  } else {
    timestampColumns = [];
  }

  let tooltipText = "";
  let buttonDisabled = true;
  $: if ($metricsInternalRep.getMetricKey("from") === "") {
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
    if (measures.length > 0 || dimensions.length > 0) {
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

<Tooltip location="right" alignment="middle" distance={5}>
  <button
    disabled={buttonDisabled}
    on:click={openModelIfNeeded}
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
        `}>quick metrics</button
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
