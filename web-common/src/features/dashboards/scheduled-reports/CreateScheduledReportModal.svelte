<script lang="ts">
  import { page } from "$app/stores";
  import { createAdminServiceCreateReport } from "@rilldata/web-admin/client";
  import Dialog from "@rilldata/web-common/components/dialog-v2/Dialog.svelte";
  import FormItemTimePicker from "@rilldata/web-common/components/forms/FormItemTimePicker.svelte";
  import { notifications } from "@rilldata/web-common/components/notifications";
  import {
    getRuntimeServiceListResourcesQueryKey,
    V1ExportFormat,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { createEventDispatcher } from "svelte";
  import { createForm } from "svelte-forms-lib";
  import * as yup from "yup";
  import { Button } from "../../../components/button";
  import FormItemInput from "../../../components/forms/FormItemInput.svelte";
  import FormItemSelect from "../../../components/forms/FormItemSelect.svelte";
  import RecipientsFormElement from "./RecipientsFormElement.svelte";
  import {
    convertToCron,
    getNextQuarterHour,
    getTimeIn24FormatFromDate,
    getTodaysDayOfWeek,
  } from "./time-utils";

  export let queryName: string;
  export let queryArgsJson: string;
  export let open: boolean;

  $: organization = $page.params.organization;
  $: project = $page.params.project;

  const createReport = createAdminServiceCreateReport();
  const dispatch = createEventDispatcher();
  const queryClient = useQueryClient();
  const queryArgs = JSON.parse(queryArgsJson);
  // console.log("queryArgs", queryArgs);

  const { form, errors, handleSubmit, isSubmitting } = createForm({
    initialValues: {
      title: "",
      frequency: "Weekly",
      dayOfWeek: getTodaysDayOfWeek(),
      timeOfDay: getTimeIn24FormatFromDate(getNextQuarterHour()),
      exportFormat: V1ExportFormat.EXPORT_FORMAT_CSV,
      exportLimit: "",
      recipients: [],
    },
    validationSchema: yup.object({
      title: yup.string().required("Required"),
      recipients: yup.array().of(yup.string()).min(1, "Required"),
    }),
    onSubmit: async (values) => {
      const refreshCron = convertToCron(
        values.frequency,
        values.dayOfWeek,
        values.timeOfDay
      );
      try {
        await $createReport.mutateAsync({
          organization,
          project,
          data: {
            options: {
              title: values.title,
              refreshCron: refreshCron, // for testing: "* * * * *"
              queryName: queryName,
              queryArgsJson: queryArgsJson,
              exportLimit: values.exportLimit || undefined,
              exportFormat: values.exportFormat,
              openProjectSubpath: `/${queryArgs.metricsViewName}`, // TODO: serialize the report parameters into the `?state` URL param
              recipients: values.recipients,
            },
          },
        });
        queryClient.invalidateQueries(
          getRuntimeServiceListResourcesQueryKey($runtime.instanceId)
        );
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
    <div class="flex gap-x-2">
      <FormItemSelect
        bind:value={$form["frequency"]}
        id="frequency"
        label="Frequency"
        options={["Daily", "Weekdays", "Weekly"].map((frequency) => ({
          value: frequency,
        }))}
      />
      {#if $form["frequency"] === "Weekly"}
        <FormItemSelect
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
      <FormItemTimePicker
        bind:value={$form["timeOfDay"]}
        id="timeOfDay"
        label="Time"
      />
    </div>
    <FormItemSelect
      bind:value={$form["exportFormat"]}
      id="exportFormat"
      label="Format"
      options={[
        { value: V1ExportFormat.EXPORT_FORMAT_CSV, label: "CSV" },
        { value: V1ExportFormat.EXPORT_FORMAT_PARQUET, label: "Parquet" },
        { value: V1ExportFormat.EXPORT_FORMAT_XLSX, label: "Excel" },
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
      error={$errors["recipients"]}
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
      <Button type="secondary" on:click={() => dispatch("close")}>Cancel</Button
      >
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
