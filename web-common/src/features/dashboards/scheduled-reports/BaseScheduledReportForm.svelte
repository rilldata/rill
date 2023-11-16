<script lang="ts">
  import TimePicker from "@rilldata/web-common/components/forms/TimePicker.svelte";
  import { V1ExportFormat } from "@rilldata/web-common/runtime-client";
  import { createForm } from "svelte-forms-lib";
  import * as yup from "yup";
  import InputV2 from "../../../components/forms/InputV2.svelte";
  import Select from "../../../components/forms/Select.svelte";
  import {
    getAbbreviationForIANA,
    getLocalIANA,
    getUTCIANA,
  } from "../../../lib/time/timezone";
  import { getStateManagers } from "../state-managers/state-managers";
  import RecipientsList from "./RecipientsList.svelte";

  export let formId: string;
  export let formState: any; // svelte-forms-lib's FormState

  const { form, errors, handleSubmit } = formState;

  const userLocalIANA = getLocalIANA();
  const ctx = getStateManagers();
  const dashboardStore = ctx.dashboardStore;
  $: dashboardTimeZone = $dashboardStore?.selectedTimezone ?? "";
  const UTCIana = getUTCIANA();

  // This form-within-a-form is used to add recipients to the parent form
  const {
    form: newRecipientForm,
    errors: newRecipientErrors,
    handleSubmit: newRecipientHandleSubmit,
  } = createForm({
    initialValues: {
      newRecipient: "",
    },
    validationSchema: yup.object({
      newRecipient: yup.string().email("Invalid email"),
    }),
    onSubmit: (values) => {
      if (values.newRecipient) {
        $form["recipients"] = $form["recipients"].concat(values.newRecipient);
      }
      $newRecipientForm.newRecipient = "";
    },
  });
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
      options={[dashboardTimeZone, userLocalIANA, UTCIana]
        // Remove duplicates when dashboardTimeZone is already covered by userLocalIANA or UTCIana
        .filter((z, i, self) => {
          return self.indexOf(z) === i;
        })
        // Add labels
        .map((z) => {
          let label = getAbbreviationForIANA(new Date(), z);
          if (z === userLocalIANA) {
            label += " (local time)";
          }
          return {
            value: z,
            label: label,
          };
        })}
    />
  </div>
  <Select
    bind:value={$form["exportFormat"]}
    id="exportFormat"
    label="Format"
    options={[
      { value: V1ExportFormat.EXPORT_FORMAT_CSV, label: "CSV" },
      { value: V1ExportFormat.EXPORT_FORMAT_PARQUET, label: "Parquet" },
      { value: V1ExportFormat.EXPORT_FORMAT_XLSX, label: "Excel" },
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
  <div class="flex flex-col gap-y-2">
    <form
      autocomplete="off"
      id="add-recipient-form"
      on:submit|preventDefault={newRecipientHandleSubmit}
    >
      <InputV2
        bind:value={$newRecipientForm["newRecipient"]}
        error={$newRecipientErrors["newRecipient"]}
        hint="Recipients may receive different views based on the project's security policies.
           Recipients without access to the project will not be able to view the report."
        id="newRecipient"
        label="Recipients"
        placeholder="Add an email address"
      />
    </form>
    <RecipientsList bind:recipients={$form["recipients"]} />
  </div>
</form>
