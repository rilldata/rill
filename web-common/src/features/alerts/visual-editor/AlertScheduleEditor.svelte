<script lang="ts">
  import Label from "@rilldata/web-common/components/forms/Label.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import TimePicker from "@rilldata/web-common/components/forms/TimePicker.svelte";
  import { makeTimeZoneOptions } from "@rilldata/web-common/features/scheduled-reports/time-utils";

  export let refreshWhenDataRefreshes: boolean;
  export let frequency: string;
  export let dayOfWeek: string;
  export let timeOfDay: string;
  export let dayOfMonth: string;
  export let timeZone: string;
  export let availableTimeZones: string[] | undefined = undefined;

  $: timeZoneOptions = makeTimeZoneOptions(availableTimeZones);

  const frequencyOptions = ["Daily", "Weekdays", "Weekly", "Monthly"].map(
    (f) => ({
      value: f,
      label: f,
    }),
  );

  const dayOfWeekOptions = [
    "Monday",
    "Tuesday",
    "Wednesday",
    "Thursday",
    "Friday",
    "Saturday",
    "Sunday",
  ].map((d) => ({
    value: d,
    label: d,
  }));
</script>

<div class="flex flex-col gap-y-3">
  <div class="flex items-center gap-x-2">
    <Switch
      bind:checked={refreshWhenDataRefreshes}
      id="refresh-when-data-refreshes"
      medium
    />
    <Label
      for="refresh-when-data-refreshes"
      class="font-medium text-fg-secondary text-sm"
    >
      Trigger whenever data refreshes
    </Label>
  </div>

  {#if !refreshWhenDataRefreshes}
    <div class="flex flex-col gap-y-2 pt-2">
      <Label class="text-sm text-fg-muted">Custom schedule</Label>

      <div class="grid grid-cols-2 gap-2">
        <Select
          bind:value={frequency}
          id="schedule-frequency"
          label="Frequency"
          options={frequencyOptions}
        />

        {#if frequency === "Weekly"}
          <Select
            bind:value={dayOfWeek}
            id="schedule-day-of-week"
            label="Day"
            options={dayOfWeekOptions}
          />
        {/if}

        {#if frequency === "Monthly"}
          <Select
            bind:value={dayOfMonth}
            id="schedule-day-of-month"
            label="Day"
            options={[{ value: "1", label: "First day" }]}
            disabled
          />
        {/if}

        <TimePicker bind:value={timeOfDay} id="schedule-time" label="Time" />

        <Select
          bind:value={timeZone}
          id="schedule-timezone"
          label="Time zone"
          options={timeZoneOptions}
        />
      </div>
    </div>
  {/if}
</div>
