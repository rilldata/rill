<script lang="ts">
  import { page } from "$app/stores";
  import { createAdminServiceCreateReport } from "@rilldata/web-admin/client";
  import Dialog from "@rilldata/web-common/components/dialog-v2/Dialog.svelte";
  import { notifications } from "@rilldata/web-common/components/notifications";
  import { V1ExportFormat } from "@rilldata/web-common/runtime-client";
  import { createEventDispatcher } from "svelte";
  import { createForm } from "svelte-forms-lib";
  import { Button } from "../../../components/button";
  import FormItemInput from "../../../components/forms/FormItemInput.svelte";
  import FormItemSelect from "../../../components/forms/FormItemSelect.svelte";
  import RecipientsFormElement from "./RecipientsFormElement.svelte";

  export let queryName: string;
  export let queryArgsJson: string;
  export let open: boolean;

  $: organization = $page.params.organization;
  $: project = $page.params.project;

  const createReport = createAdminServiceCreateReport();
  const dispatch = createEventDispatcher();

  const { form, errors, handleSubmit, isSubmitting } = createForm({
    initialValues: {
      title: "",
      // firstRunAtDate: new Date().toISOString().split("T")[0], // Today's date
      // firstRunAtTime: formatTime(getNextQuarterHour()), // Next quarter hour
      // firstRunAtTimezone: "UTC",
      refreshCron: "* * * * *",
      exportFormat: V1ExportFormat.EXPORT_FORMAT_CSV,
      exportLimit: "",
      recipients: [],
    },
    // This isn't showing issues
    // validationSchema: yup.object({
    //   title: yup.string().required("Required"),
    //   firstRunAtDate: yup.string().required("Required"),
    //   firstRunAtTime: yup.string().required("Required"),
    //   firstRunAtTimezone: yup.string().required("Required"),
    //   refreshCron: yup.string().required("Required"),
    //   recipients: yup.string().required("Required"),
    // }),
    onSubmit: async (values) => {
      try {
        await $createReport.mutateAsync({
          organization,
          project,
          data: {
            options: {
              title: values.title,
              refreshCron: values.refreshCron,
              queryName: queryName,
              queryArgsJson: queryArgsJson,
              exportLimit: values.exportLimit || undefined,
              exportFormat: values.exportFormat,
              openProjectSubpath: "/-/reports", // It'd be nice for this to be the specific report's path, but we don't have that data at request time
              recipients: values.recipients,
            },
          },
        });
        dispatch("close");
        notifications.send({
          message: "Report created",
          type: "success",
        });
      } catch (e) {
        // showing error below
      }
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
      bind:value={$form["title"]}
      error={$errors["title"]}
      id="title"
      label="Report title"
      placeholder="My report"
    />
    <!-- Hide while backend doesn't support a "firstRunAt" time -->
    <!-- <div class="flex items-end gap-x-2 w-full">
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
    </div> -->
    <FormItemSelect
      bind:value={$form["refreshCron"]}
      id="refreshCron"
      label="Frequency"
      options={["Daily", "Weekdays", "Weekly", "Monthly"]}
    />
    <FormItemSelect
      bind:value={$form["exportFormat"]}
      id="exportFormat"
      label="Format"
      options={[
        V1ExportFormat.EXPORT_FORMAT_CSV,
        V1ExportFormat.EXPORT_FORMAT_PARQUET,
        V1ExportFormat.EXPORT_FORMAT_XLSX,
      ]}
    />
    <FormItemInput
      bind:value={$form["exportLimit"]}
      error={$errors["exportLimit"]}
      id="exportLimit"
      label="Row limit"
      placeholder="1000"
      optional
    />
    <RecipientsFormElement
      bind:recipients={$form["recipients"]}
      {organization}
      {project}
    />
  </form>
  <svelte:fragment slot="footer">
    <div class="flex items-center gap-x-2">
      {#if $createReport.isError}
        <div class="text-red-500">{$createReport.error.message}</div>
      {/if}
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
