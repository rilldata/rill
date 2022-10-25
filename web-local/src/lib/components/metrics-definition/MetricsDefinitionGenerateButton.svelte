<script lang="ts">
  import { useQueryClient } from "@sveltestack/svelte-query";
  import { getContext } from "svelte";
  import type { DerivedModelStore } from "../../application-state-stores/model-stores";
  import { getDimensionsByMetricsId } from "../../redux-store/dimension-definition/dimension-definition-readables";
  import { getMeasuresByMetricsId } from "../../redux-store/measure-definition/measure-definition-readables";
  import {
    generateMeasuresAndDimensionsApi,
    updateMetricsDefsWrapperApi,
  } from "../../redux-store/metrics-definition/metrics-definition-apis";
  import { getMetricsDefReadableById } from "../../redux-store/metrics-definition/metrics-definition-readables";
  import { selectTimestampColumnFromProfileEntity } from "../../redux-store/source/source-selectors";
  import { store } from "../../redux-store/store-root";
  import { invalidateMetricsView } from "../../svelte-query/queries/metrics-views/invalidation";
  import type { ProfileColumn } from "../../types";
  import Tooltip from "../tooltip/Tooltip.svelte";
  import TooltipContent from "../tooltip/TooltipContent.svelte";
  import QuickMetricsModal from "./QuickMetricsModal.svelte";

  $: selectedMetricsDef = getMetricsDefReadableById(metricsDefId);
  $: selectedDimensions = getDimensionsByMetricsId(metricsDefId);
  $: selectedMeasures = getMeasuresByMetricsId(metricsDefId);

  export let metricsDefId: string;

  const queryClient = useQueryClient();

  async function handleGenerateClick() {
    await store.dispatch(generateMeasuresAndDimensionsApi(metricsDefId));
    if (!$selectedMetricsDef?.timeDimension && timestampColumns.length > 0) {
      // select the first available timestamp column if one has not been
      // selected and there are some available
      store.dispatch(
        updateMetricsDefsWrapperApi({
          id: metricsDefId,
          changes: { timeDimension: timestampColumns[0].name },
        })
      );
    }
    invalidateMetricsView(queryClient, metricsDefId);
    // In `svelte-query/totals.ts`, in the `invalidateMetricsViewData()` function, we use `refetchQueries` where we should probably use `invalidateQueries`.
    // We should make that change, but it has a wide surface area, so we need to take the time to QA it properly. In the meantime, we need to remove old
    // queryKeys from the queryCache, so they don't get refetched and generate 500 errors.
    queryClient.removeQueries([`v1/metrics-view/toplist`, metricsDefId], {
      exact: false,
    });
    closeModal();
  }

  const derivedModelStore = getContext(
    "rill:app:derived-model-store"
  ) as DerivedModelStore;

  let timestampColumns: Array<ProfileColumn>;

  $: if ($selectedMetricsDef?.sourceModelId && $derivedModelStore?.entities) {
    timestampColumns = selectTimestampColumnFromProfileEntity(
      $derivedModelStore?.entities.find(
        (model) => model.id === $selectedMetricsDef.sourceModelId
      )
    );
  } else {
    timestampColumns = [];
  }

  let tooltipText = "";
  let buttonDisabled = true;
  $: if ($selectedMetricsDef?.sourceModelId === undefined) {
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
    if ($selectedDimensions.length > 0 || $selectedMeasures.length > 0) {
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
