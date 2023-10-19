<script lang="ts">
  import { page } from "$app/stores";
  import Dialog from "@rilldata/web-common/components/dialog-v2/Dialog.svelte";
  import { createForm } from "svelte-forms-lib";
  import * as yup from "yup";
  import { Button } from "../../../components/button";
  import FormItemDatePicker from "../../../components/forms/FormItemDatePicker.svelte";
  import FormItemInput from "../../../components/forms/FormItemInput.svelte";
  import FormItemSelect from "../../../components/forms/FormItemSelect.svelte";
  import FormItemTimePicker from "../../../components/forms/FormItemTimePicker.svelte";
  import {
    formatTime,
    getNextQuarterHour,
  } from "../../../components/forms/time-utils";
  import RecipientsFormElement from "./RecipientsFormElement.svelte";

  export let metricViewName: string;
  export let open: boolean;

  $: organization = $page.params.organization;
  $: project = $page.params.project;

  const { form, errors, handleSubmit, isSubmitting } = createForm({
    initialValues: {
      reportName: "",
      firstRunAtDate: new Date().toISOString().split("T")[0], // Today's date
      firstRunAtTime: formatTime(getNextQuarterHour()), // Next quarter hour
      firstRunAtTimezone: "UTC",
      frequency: "Daily",
      limit: "",
      recipients: [],
    },
    validationSchema: yup.object({
      reportName: yup.string().required("Required"),
      firstRunAtDate: yup.string().required("Required"),
      firstRunAtTime: yup.string().required("Required"),
      firstRunAtTimezone: yup.string().required("Required"),
      frequency: yup.string().required("Required"),
      recipients: yup.string().required("Required"),
    }),
    onSubmit: async (values) => {
      console.log(`Submit form for ${metricViewName}`, values);
    },
  });
</script>

<Dialog {open} on:close>
  <svelte:fragment slot="title">Schedule report</svelte:fragment>
  <form
    on:submit|preventDefault={handleSubmit}
    id="create-scheduled-report-form"
    autocomplete="off"
    class="flex flex-col gap-y-6"
    slot="body"
  >
    <span>Email recurring exports to recipients.</span>
    <FormItemInput
      bind:value={$form["reportName"]}
      error={$errors["reportName"]}
      id="reportName"
      label="Report name"
      placeholder="My report"
    />
    <!-- error={$errors["firstRunAt"]} -->
    <div class="flex items-end gap-x-2 w-full">
      <FormItemDatePicker
        bind:value={$form["firstRunAtDate"]}
        id="firstRunAtDate"
        label="First run at"
      />
      <FormItemTimePicker
        bind:value={$form["firstRunAtTime"]}
        id="firstRunAtTime"
        label=""
      />
      <FormItemSelect
        bind:value={$form["firstRunAtTimezone"]}
        id="firstRunAtTimezone"
        label=""
        options={["UTC"]}
      />
    </div>
    <FormItemSelect
      bind:value={$form["frequency"]}
      id="frequency"
      label="Frequency"
      options={["Daily", "Weekdays", "Weekly", "Monthly"]}
    />
    <FormItemInput
      bind:value={$form["limit"]}
      error={$errors["limit"]}
      id="limit"
      label="Row limit"
      placeholder="1000 (rows)"
      optional
    />
    <RecipientsFormElement
      bind:recipients={$form["recipients"]}
      {organization}
      {project}
    />
  </form>
  <svelte:fragment slot="footer">
    <div class="flex gap-x-2">
      <div class="grow" />
      <Button type="secondary" on:click={close}>Cancel</Button>
      <Button
        type="primary"
        submitForm
        form="create-scheduled-report-form"
        disabled={$isSubmitting}
      >
        Create
      </Button>
    </div>
  </svelte:fragment>
</Dialog>
