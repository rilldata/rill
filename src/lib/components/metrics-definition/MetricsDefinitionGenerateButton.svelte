<script lang="ts">
  import { getContext } from "svelte";

  import type { DerivedModelStore } from "$lib/application-state-stores/model-stores";
  import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
  import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";
  import { getDimensionsByMetricsId } from "$lib/redux-store/dimension-definition/dimension-definition-readables";
  import { getMeasuresByMetricsId } from "$lib/redux-store/measure-definition/measure-definition-readables";
  import {
    generateMeasuresAndDimensionsApi,
    updateMetricsDefsWrapperApi,
  } from "$lib/redux-store/metrics-definition/metrics-definition-apis";
  import { getMetricsDefReadableById } from "$lib/redux-store/metrics-definition/metrics-definition-readables";
  import { selectTimestampColumnFromProfileEntity } from "$lib/redux-store/source/source-selectors";
  import { store } from "$lib/redux-store/store-root";
  import type { ProfileColumn } from "$lib/types";
  import QuickMetricsModal from "./QuickMetricsModal.svelte";

  $: selectedMetricsDef = getMetricsDefReadableById(metricsDefId);
  $: selectedDimensions = getDimensionsByMetricsId(metricsDefId);
  $: selectedMeasures = getMeasuresByMetricsId(metricsDefId);

  export let metricsDefId: string;

  function handleGenerateClick() {
    store.dispatch(generateMeasuresAndDimensionsApi(metricsDefId));
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
    on:replace-metrics={handleGenerateClick}
  />
{/if}
