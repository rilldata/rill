<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { Button } from "../../../components/button";

  import { getISOStringFromDate } from "./time-range-utils";

  export let min: string;
  export let max: string;
  export let initialStartDate: string;
  export let initialEndDate: string;
  export let validateCustomTimeRange: (start: string, end: string) => string;

  const dispatch = createEventDispatcher();

  let start: string = initialStartDate;
  let end: string = initialEndDate;

  $: error = validateCustomTimeRange(start, end);
  $: disabled = !start || !end || !!error;

  function applyCustomTimeRange() {
    // Currently, we assume UTC
    dispatch("apply", {
      startDate: getISOStringFromDate(start),
      endDate: getISOStringFromDate(end),
    });
  }

  let labelClasses = "font-semibold text-[10px]";
</script>

<form
  id="custom-time-range-form"
  class="flex flex-col gap-y-3 mt-3 mb-1 px-3"
  on:submit|preventDefault={applyCustomTimeRange}
>
  <div class="flex flex-col gap-y-1">
    <label for="start-date" class={labelClasses}>Start date</label>
    <input
      bind:value={start}
      on:blur={() => dispatch("close-calendar")}
      type="date"
      id="start-date"
      name="start-date"
      {min}
      {max}
      class="cursor-pointer"
    />
  </div>
  <div class="flex flex-col gap-y-1">
    <label for="end-date" class={labelClasses}>End date</label>

    <input
      bind:value={end}
      on:blur={() => dispatch("close-calendar")}
      type="date"
      id="end-date"
      name="end-date"
      {min}
      {max}
    />
  </div>
  <div class="flex mt-3 items-center">
    {#if error}
      <div style:font-size="11px" class="text-red-600">
        {error}
      </div>
    {/if}
    <div class="flex-grow" />
    <Button type="primary" submitForm form="custom-time-range-form" {disabled}>
      Apply
    </Button>
  </div>
</form>
