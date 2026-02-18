<script lang="ts">
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import TimePicker from "@rilldata/web-common/components/forms/TimePicker.svelte";
  import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors.ts";
  import {
    getInitialScheduleFormValues,
    makeTimeZoneOptions,
    ReportFrequency,
  } from "@rilldata/web-common/features/scheduled-reports/time-utils.ts";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store.ts";
  import type { Readable } from "svelte/store";

  export let data: Readable<ReturnType<typeof getInitialScheduleFormValues>>;
  export let exploreName: string;

  $: ({ instanceId } = $runtime);

  // Pull the time zone options from the dashboard's spec
  $: exploreSpec = useExploreValidSpec(instanceId, exploreName);
  $: availableTimeZones = exploreSpec.data?.explore?.timeZones;
  $: timeZoneOptions = makeTimeZoneOptions(availableTimeZones);
</script>

<div class="flex gap-x-1">
  <Select
    bind:value={$data["frequency"]}
    id="frequency"
    label="Frequency"
    options={["Daily", "Weekdays", "Weekly", "Monthly"].map((frequency) => ({
      value: frequency,
      label: frequency,
    }))}
  />
  {#if $data["frequency"] === ReportFrequency.Weekly}
    <Select
      bind:value={$data["dayOfWeek"]}
      id="dayOfWeek"
      label="Day"
      options={[
        "Monday",
        "Tuesday",
        "Wednesday",
        "Thursday",
        "Friday",
        "Saturday",
        "Sunday",
      ].map((day) => ({
        value: day,
        label: day,
      }))}
    />
  {/if}
  {#if $data["frequency"] === ReportFrequency.Monthly}
    <Select
      value={"1"}
      id="dayOfMonth"
      label="Day"
      options={[{ value: "1", label: "First day" }]}
      disabled
    />
  {/if}
  <TimePicker bind:value={$data["timeOfDay"]} id="timeOfDay" label="Time" />
  <Select
    bind:value={$data["timeZone"]}
    id="timeZone"
    label="Time zone"
    options={timeZoneOptions}
  />
</div>
