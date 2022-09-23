<script lang="ts">
  import TimestampIcon from "$lib/components/icons/TimestampType.svelte";
  import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
  import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";

  import { store } from "$lib/redux-store/store-root";
  import { getContext } from "svelte";
  import type { DerivedModelStore } from "$lib/application-state-stores/model-stores";
  import type { ProfileColumn } from "$lib/types";
  import { updateMetricsDefsWrapperApi } from "$lib/redux-store/metrics-definition/metrics-definition-apis";
  import { getMetricsDefReadableById } from "$lib/redux-store/metrics-definition/metrics-definition-readables";
  import { selectTimestampColumnFromProfileEntity } from "$lib/redux-store/source/source-selectors";

  export let metricsDefId;

  $: selectedMetricsDef = getMetricsDefReadableById(metricsDefId);

  $: timeColumnSelectedValue =
    $selectedMetricsDef?.timeDimension || "__DEFAULT_VALUE__";

  const derivedModelStore = getContext(
    "rill:app:derived-model-store"
  ) as DerivedModelStore;

  let derivedModelColumns: Array<ProfileColumn>;
  $: if ($selectedMetricsDef?.sourceModelId && $derivedModelStore?.entities) {
    derivedModelColumns = selectTimestampColumnFromProfileEntity(
      $derivedModelStore?.entities.find(
        (model) => model.id === $selectedMetricsDef.sourceModelId
      )
    );
  } else {
    derivedModelColumns = [];
  }

  function updateMetricsDefinitionHandler(evt: Event) {
    store.dispatch(
      updateMetricsDefsWrapperApi({
        id: metricsDefId,
        changes: { timeDimension: (<HTMLSelectElement>evt.target).value },
      })
    );
  }

  let tooltipText = "";
  let dropdownDisabled = true;
  $: if ($selectedMetricsDef?.sourceModelId === undefined) {
    tooltipText = "select a model before selecting a timestamp column";
    dropdownDisabled = true;
  } else if (derivedModelColumns.length === 0) {
    tooltipText = "the selected model has no timestamp columns";
    dropdownDisabled = true;
  } else {
    tooltipText = undefined;
    dropdownDisabled = false;
  }
</script>

<div class="flex items-center">
  <div class="flex items-center gap-x-2" style="width:9em">
    <TimestampIcon size="16px" /> timestamp
  </div>
  <div>
    <Tooltip
      location="right"
      alignment="middle"
      distance={16}
      suppress={tooltipText === undefined}
    >
      <select
        class="italic hover:bg-gray-100 rounded border border-6 border-transparent hover:font-bold hover:border-gray-100"
        style="background-color: #FFF; width:18em;"
        value={timeColumnSelectedValue}
        on:change={updateMetricsDefinitionHandler}
        disabled={dropdownDisabled}
      >
        <option value="__DEFAULT_VALUE__" disabled selected hidden
          >select a timestamp...</option
        >
        {#each derivedModelColumns as column}
          <option value={column.name}>{column.name}</option>
        {/each}
      </select>

      <TooltipContent slot="tooltip-content">
        {tooltipText}
      </TooltipContent>
    </Tooltip>
  </div>
</div>
