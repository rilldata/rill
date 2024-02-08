<script lang="ts">
  import TimePicker from "@rilldata/web-common/components/forms/TimePicker.svelte";
  import RecipientsInputArray from "@rilldata/web-common/features/scheduled-reports/RecipientsInputArray.svelte";
  import { V1ExportFormat } from "@rilldata/web-common/runtime-client";
  import InputV2 from "../../components/forms/InputV2.svelte";
  import Select from "../../components/forms/Select.svelte";
  import { runtime } from "../../runtime-client/runtime-store";
  import { useDashboard } from "../dashboards/selectors";
  import { makeTimeZoneOptions } from "./time-utils";

  export let formId: string;
  export let formState: any; // svelte-forms-lib's FormState
  export let metricsViewName: string;

  const { form, errors, handleSubmit } = formState;

  // Pull the time zone options from the dashboard's spec
  $: dashboard = useDashboard($runtime.instanceId, metricsViewName);
  $: availableTimeZones =
    $dashboard.data?.metricsView?.spec?.availableTimeZones;
  $: timeZoneOptions = makeTimeZoneOptions(availableTimeZones);
</script>

<form
  autocomplete="off"
  class="flex flex-col gap-y-6"
  id={formId}
  on:submit|preventDefault={handleSubmit}
>
  <span>Email recurring exports to recipients.</span>
  <InputV2
    bind:value={$form["title"]}
    error={$errors["title"]}
    id="title"
    label="Report title"
    placeholder="My report"
  />
  <div class="flex gap-x-2">
    <Select
      bind:value={$form["frequency"]}
      id="frequency"
      label="Frequency"
      options={["Daily", "Weekdays", "Weekly"].map((frequency) => ({
        value: frequency,
      }))}
    />
    {#if $form["frequency"] === "Weekly"}
      <Select
        bind:value={$form["dayOfWeek"]}
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
        }))}
      />
    {/if}
    <TimePicker bind:value={$form["timeOfDay"]} id="timeOfDay" label="Time" />
    <Select
      bind:value={$form["timeZone"]}
      id="timeZone"
      label="Time zone"
      options={timeZoneOptions}
    />
  </div>
  <Select
    bind:value={$form["exportFormat"]}
    id="exportFormat"
    label="Format"
    options={[
      { value: V1ExportFormat.EXPORT_FORMAT_CSV, label: "CSV" },
      { value: V1ExportFormat.EXPORT_FORMAT_PARQUET, label: "Parquet" },
      { value: V1ExportFormat.EXPORT_FORMAT_XLSX, label: "XLSX" },
    ]}
  />
  <InputV2
    bind:value={$form["exportLimit"]}
    error={$errors["exportLimit"]}
    id="exportLimit"
    label="Row limit"
    optional
    placeholder="1000"
  />
  <RecipientsInputArray
    {formState}
    hint="Recipients will receive different views based on their security policy.
        Recipients without project access can't view the report."
  />
</form>
