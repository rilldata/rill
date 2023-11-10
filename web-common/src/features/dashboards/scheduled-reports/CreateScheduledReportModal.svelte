<script lang="ts">
  import { page } from "$app/stores";
  import { createAdminServiceCreateReport } from "@rilldata/web-admin/client";
  import Dialog from "@rilldata/web-common/components/dialog-v2/Dialog.svelte";
  import TimePicker from "@rilldata/web-common/components/forms/TimePicker.svelte";
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
  import InputV2 from "../../../components/forms/InputV2.svelte";
  import Select from "../../../components/forms/Select.svelte";
  import {
    getAbbreviationForIANA,
    getLocalIANA,
    getUTCIANA,
  } from "../../../lib/time/timezone";
  import RecipientsList from "./RecipientsList.svelte";
  import {
    convertToCron,
    getNextQuarterHour,
    getTimeIn24FormatFromDateTime,
    getTodaysDayOfWeek,
  } from "./time-utils";

  export let queryName: string;
  export let queryArgsJson: string;
  export let dashboardTimeZone: string;
  export let open: boolean;

  $: organization = $page.params.organization;
  $: project = $page.params.project;

  const createReport = createAdminServiceCreateReport();
  const dispatch = createEventDispatcher();
  const queryClient = useQueryClient();
  const queryArgs = JSON.parse(queryArgsJson);

  const userLocalIANA = getLocalIANA();
  const UTCIana = getUTCIANA();

  // TODO: a better approach will be to use the queryArgs to craft the right state object
  const dashState = new URLSearchParams(window.location.search).get("state");

  const { form, errors, handleSubmit, isSubmitting } = createForm({
    initialValues: {
      title: "",
      frequency: "Weekly",
      dayOfWeek: getTodaysDayOfWeek(),
      timeOfDay: getTimeIn24FormatFromDateTime(getNextQuarterHour()),
      timeZone: dashboardTimeZone || userLocalIANA,
      exportFormat: V1ExportFormat.EXPORT_FORMAT_CSV,
      exportLimit: "",
      recipients: [] as string[],
    },
    validationSchema: yup.object({
      title: yup.string().required("Required"),
      recipients: yup.array().min(1, "Required"),
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
              refreshTimeZone: values.timeZone,
              queryName: queryName,
              queryArgsJson: queryArgsJson,
              exportLimit: values.exportLimit || undefined,
              exportFormat: values.exportFormat,
              openProjectSubpath: `/${queryArgs.metricsViewName}?state=${dashState}`,
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

<Dialog {open}>
  <svelte:fragment slot="title">Schedule report</svelte:fragment>
  <form
    on:submit|preventDefault={handleSubmit}
    id="create-scheduled-report-form"
    autocomplete="off"
    class="flex flex-col gap-y-6"
    slot="body"
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
      placeholder="1000"
      optional
    />
    <div class="flex flex-col gap-y-2">
      <form
        on:submit|preventDefault={newRecipientHandleSubmit}
        id="add-recipient-form"
        autocomplete="off"
      >
        <InputV2
          bind:value={$newRecipientForm["newRecipient"]}
          error={$newRecipientErrors["newRecipient"]}
          id="newRecipient"
          label="Recipients"
          placeholder="Add an email address"
          hint="Recipients may receive different views based on the project's security policies.
           Recipients without access to the project will not be able to view the report."
        />
      </form>
      <RecipientsList bind:recipients={$form["recipients"]} />
    </div>
  </form>
  <svelte:fragment slot="footer">
    <div class="flex items-center gap-x-2 mt-2">
      {#if $createReport.isError}
        <div class="text-red-500">{$createReport.error.message}</div>
      {/if}
      <div class="grow" />
      <Button type="secondary" on:click={() => dispatch("close")}>
        Cancel
      </Button>
      <Button
        type="primary"
        submitForm
        form="create-scheduled-report-form"
        disabled={$isSubmitting || $form["recipients"].length === 0}
      >
        Create
      </Button>
    </div>
  </svelte:fragment>
</Dialog>
