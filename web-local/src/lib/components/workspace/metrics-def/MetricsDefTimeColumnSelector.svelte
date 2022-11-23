<script lang="ts">
  import { useRuntimeServiceGetCatalogEntry } from "@rilldata/web-common/runtime-client";
  import TimestampIcon from "../../icons/TimestampType.svelte";
  import Tooltip from "../../tooltip/Tooltip.svelte";
  import TooltipContent from "../../tooltip/TooltipContent.svelte";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { TIMESTAMPS } from "@rilldata/web-local/lib/duckdb-data-types";

  export let metricsInternalRep;

  $: instanceId = $runtimeStore.instanceId;

  $: model = $metricsInternalRep.getMetricKey("from");
  $: getModel = useRuntimeServiceGetCatalogEntry(instanceId, model);
  $: selectedModel = $getModel.data?.entry?.model;

  $: timeColumnSelectedValue =
    $metricsInternalRep.getMetricKey("timeseries") || "__DEFAULT_VALUE__";

  let timestampColumns: Array<string>;
  $: if (selectedModel) {
    const selectedMetricsDefModelProfile = selectedModel?.schema?.fields ?? [];
    timestampColumns = selectedMetricsDefModelProfile
      .filter((column) => TIMESTAMPS.has(column.type.code as string))
      .map((column) => column.name);
  } else {
    timestampColumns = [];
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
  } else if (timestampColumns.length === 0) {
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
        {#each timestampColumns as column}
          <option value={column}>{column}</option>
        {/each}
      </select>

      <TooltipContent slot="tooltip-content">
        {tooltipText}
      </TooltipContent>
    </Tooltip>
  </div>
</div>
