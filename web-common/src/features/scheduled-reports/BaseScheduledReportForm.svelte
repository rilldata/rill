<script lang="ts">
  import TimePicker from "@rilldata/web-common/components/forms/TimePicker.svelte";
  import { V1ExportFormat } from "@rilldata/web-common/runtime-client";
  import InputArray from "../../components/forms/InputArray.svelte";
  import InputV2 from "../../components/forms/InputV2.svelte";
  import Select from "../../components/forms/Select.svelte";
  import { runtime } from "../../runtime-client/runtime-store";
  import { useDashboard } from "../dashboards/selectors";
  import { makeTimeZoneOptions } from "./time-utils";

  export let formId: string;
  export let formState: any; // svelte-forms-lib's FormState
  export let metricsViewName: string;

  const { form, errors, handleSubmit } = formState;

  // There's a bug in how `svelte-forms-lib` types the `$errors` store for arrays.
  // See: https://github.com/tjinauyeung/svelte-forms-lib/issues/154#issuecomment-1087331250
  $: recipientErrors = $errors.recipients as unknown as { email: string }[];

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
  <InputArray
    id="recipients"
    label="Recipients"
    bind:values={$form["recipients"]}
    bind:errors={recipientErrors}
    accessorKey="email"
    hint="Recipients will receive different views based on their security policy.
        Recipients without project access can't view the report."
    placeholder="Enter an email address"
    addItemLabel="Add email"
    on:add-item={() => {
      $form["recipients"] = $form["recipients"].concat({ email: "" });
      recipientErrors = recipientErrors.concat({ email: "" });

      // Focus on the new input element
      setTimeout(() => {
        const input = document.getElementById(
          `recipients.${$form["recipients"].length - 1}.email`,
        );
        input?.focus();
      }, 0);
    }}
    on:remove-item={(event) => {
      const index = event.detail.index;
      $form["recipients"] = $form["recipients"].filter((r, i) => i !== index);
      recipientErrors = recipientErrors.filter((r, i) => i !== index);
    }}
  />
</form>
