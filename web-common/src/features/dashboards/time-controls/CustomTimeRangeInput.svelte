<script lang="ts">
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { createEventDispatcher } from "svelte";
  import { Button } from "../../../components/button";
  import { useRuntimeServiceGetTimeRangeSummary } from "../../../runtime-client";
  import { useMetaQuery } from "../selectors";

  export let metricViewName: string;

  const dispatch = createEventDispatcher();

  let start: string;
  let end: string;

  $: disabled = !start || !end;

  $: metaQuery = useMetaQuery($runtimeStore.instanceId, metricViewName);
  $: timeRangeQuery = useRuntimeServiceGetTimeRangeSummary(
    $runtimeStore.instanceId,
    $metaQuery.data?.model,
    { columnName: $metaQuery.data?.timeDimension },
    {
      query: {
        enabled: !!$metaQuery.data,
      },
    }
  );

  // TODO: fix time zone handling
  // temporary hack: datetime-local assumes local time, so we need to remove the time zone information
  $: min = $timeRangeQuery.data.timeRangeSummary.min.replace(/Z$/, "");
  $: max = $timeRangeQuery.data.timeRangeSummary.max.replace(/Z$/, "");

  function applyCustomTimeRange() {
    dispatch("apply", { startDate: start, endDate: end });
  }

  function handleKeydown(event: KeyboardEvent) {
    if (event.key === "Enter" && !disabled) {
      applyCustomTimeRange();
    }
  }
</script>

<form id="custom-time-range-form" class="flex flex-col gap-y-3 my-3 px-3">
  <div class="flex flex-col gap-y-1">
    <label for="start-date" style="font-size: 10px;">Start date</label>
    <input
      bind:value={start}
      type="datetime-local"
      id="start-date"
      name="start-date"
      {min}
      {max}
      on:keydown={handleKeydown}
    />
  </div>
  <div class="flex flex-col gap-y-1">
    <label for="end-date" style="font-size: 10px;">End date</label>
    <input
      bind:value={end}
      type="datetime-local"
      id="end-date"
      name="end-date"
      {min}
      {max}
      on:keydown={handleKeydown}
    />
  </div>
  <div class="flex">
    <div class="flex-grow" />
    <Button
      type="primary"
      submitForm
      form="custom-time-range-form"
      {disabled}
      on:click={applyCustomTimeRange}>Apply</Button
    >
  </div>
</form>
