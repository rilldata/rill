<script lang="ts">
  import { useRuntimeServiceGetCatalogEntry } from "@rilldata/web-common/runtime-client";
  import TimestampIcon from "../../icons/TimestampType.svelte";
  import Tooltip from "../../tooltip/Tooltip.svelte";
  import TooltipContent from "../../tooltip/TooltipContent.svelte";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";

  import { getContext } from "svelte";
  import type { DerivedModelStore } from "../../../application-state-stores/model-stores";
  import type { ProfileColumn } from "../../../types";
  import { selectTimestampColumnFromProfileEntity } from "../../../redux-store/source/source-selectors";

  export let metricsInternalRep;

  $: instanceId = $runtimeStore.instanceId;

  $: model = $metricsInternalRep.getMetricKey("from");
  $: getModel = useRuntimeServiceGetCatalogEntry(instanceId, model);
  $: selectedModel = $getModel.data?.entry?.model;

  $: timeColumnSelectedValue =
    $metricsInternalRep.getMetricKey("timeseries") || "__DEFAULT_VALUE__";

  const derivedModelStore = getContext(
    "rill:app:derived-model-store"
  ) as DerivedModelStore;

  let derivedModelColumns: Array<ProfileColumn>;
  $: if (selectedModel && $derivedModelStore?.entities) {
    derivedModelColumns = selectTimestampColumnFromProfileEntity(
      $derivedModelStore?.entities.find(
        (model) => model.id === selectedModel.name // Use model name, this is temp
      )
    );
  } else {
    derivedModelColumns = [];
  }

  function updateMetricsDefinitionHandler(evt: Event) {
    $metricsInternalRep.updateMetricKey(
      "timeseries",
      (<HTMLSelectElement>evt.target).value
    );
  }

  let tooltipText = "";
  let dropdownDisabled = true;
  $: if (selectedModel?.name === undefined) {
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
      alignment="middle"
      distance={16}
      location="right"
      suppress={tooltipText === undefined}
    >
      <select
        class="italic hover:bg-gray-100 rounded border border-6 border-transparent hover:font-bold hover:border-gray-100"
        disabled={dropdownDisabled}
        on:change={updateMetricsDefinitionHandler}
        style="background-color: #FFF; width:18em;"
        value={timeColumnSelectedValue}
      >
        <option disabled hidden selected value="__DEFAULT_VALUE__"
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
