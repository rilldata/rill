<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import TimePicker from "@rilldata/web-common/components/forms/TimePicker.svelte";
  import {
    getInitialScheduleFormValues,
    makeTimeZoneOptions,
    ReportFrequency,
  } from "@rilldata/web-common/features/scheduled-reports/time-utils.ts";
  import type { Readable } from "svelte/store";

  export let data: Readable<ReturnType<typeof getInitialScheduleFormValues>>;
  // Time zone options from the dashboard's spec (falls back to Local + UTC when undefined)
  export let availableTimeZones: string[] | undefined = undefined;

  $: timeZoneOptions = makeTimeZoneOptions(availableTimeZones);
</script>

<div class="flex gap-x-1">
  <Select
    bind:value={$data["frequency"]}
    id="frequency"
    label={m.report_form_frequency()}
    options={[
      { value: "Daily", label: m.report_form_freq_daily() },
      { value: "Weekdays", label: m.report_form_freq_weekdays() },
      { value: "Weekly", label: m.report_form_freq_weekly() },
      { value: "Monthly", label: m.report_form_freq_monthly() },
    ]}
  />
  {#if $data["frequency"] === ReportFrequency.Weekly}
    <Select
      bind:value={$data["dayOfWeek"]}
      id="dayOfWeek"
      label={m.report_form_day()}
      options={[
        { value: "Monday", label: m.report_form_day_monday() },
        { value: "Tuesday", label: m.report_form_day_tuesday() },
        { value: "Wednesday", label: m.report_form_day_wednesday() },
        { value: "Thursday", label: m.report_form_day_thursday() },
        { value: "Friday", label: m.report_form_day_friday() },
        { value: "Saturday", label: m.report_form_day_saturday() },
        { value: "Sunday", label: m.report_form_day_sunday() },
      ]}
    />
  {/if}
  {#if $data["frequency"] === ReportFrequency.Monthly}
    <Select
      value={"1"}
      id="dayOfMonth"
      label={m.report_form_day()}
      options={[{ value: "1", label: m.report_form_day_first() }]}
      disabled
    />
  {/if}
  <TimePicker
    bind:value={$data["timeOfDay"]}
    id="timeOfDay"
    label={m.report_form_time()}
  />
  <Select
    bind:value={$data["timeZone"]}
    id="timeZone"
    label={m.report_form_timezone()}
    options={timeZoneOptions}
  />
</div>
