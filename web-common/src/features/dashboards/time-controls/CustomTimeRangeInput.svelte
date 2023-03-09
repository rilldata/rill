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
  class="flex flex-col gap-y-3 mt-3 mb-1 px-3"
  id="custom-time-range-form"
  on:submit|preventDefault={applyCustomTimeRange}
>
  <div class="flex flex-col gap-y-1">
    <label class={labelClasses} for="start-date">Start date</label>
    <input
      bind:value={start}
      class="cursor-pointer"
      id="start-date"
      {max}
      {min}
      name="start-date"
      on:blur={() => dispatch("close-calendar")}
      type="date"
    />
  </div>
  <div class="flex flex-col gap-y-1">
    <label class={labelClasses} for="end-date">End date</label>

    <input
      bind:value={end}
      id="end-date"
      {max}
      {min}
      name="end-date"
      on:blur={() => dispatch("close-calendar")}
      type="date"
    />
  </div>
  <div class="flex mt-3 items-center">
    {#if error}
      <div style:font-size="11px" class="text-red-600">
        {error}
      </div>
    {/if}
    <div class="flex-grow" />
    <Button {disabled} form="custom-time-range-form" submitForm type="primary">
      Apply
    </Button>
  </div>
</form>
