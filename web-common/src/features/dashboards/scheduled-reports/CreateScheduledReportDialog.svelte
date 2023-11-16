<script lang="ts">
  import { page } from "$app/stores";
  import { createAdminServiceCreateReport } from "@rilldata/web-admin/client";
  import Dialog from "@rilldata/web-common/components/dialog-v2/Dialog.svelte";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { createEventDispatcher } from "svelte";
  import { createForm } from "svelte-forms-lib";
  import * as yup from "yup";
  import { Button } from "../../../components/button";
  import { notifications } from "../../../components/notifications";
  import { getLocalIANA } from "../../../lib/time/timezone";
  import {
    getRuntimeServiceListResourcesQueryKey,
    V1ExportFormat,
  } from "../../../runtime-client";
  import { runtime } from "../../../runtime-client/runtime-store";
  import BaseScheduledReportForm from "./BaseScheduledReportForm.svelte";
  import {
    convertToCron,
    getNextQuarterHour,
    getTimeIn24FormatFromDateTime,
    getTodaysDayOfWeek,
  } from "./time-utils";

  export let open: boolean;
  export let queryName: string;
  export let queryArgs: any;

  const createReport = createAdminServiceCreateReport();
  $: organization = $page.params.organization;
  $: project = $page.params.project;
  const dashState = new URLSearchParams(window.location.search).get("state");
  const queryClient = useQueryClient();
  const dispatch = createEventDispatcher();

  const formState = createForm({
    initialValues: {
      title: "",
      frequency: "Weekly",
      dayOfWeek: getTodaysDayOfWeek(),
      timeOfDay: getTimeIn24FormatFromDateTime(getNextQuarterHour()),
      timeZone: getLocalIANA(),
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
              queryArgsJson: JSON.stringify(queryArgs),
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

  const { isSubmitting, form } = formState;
</script>

<Dialog {open}>
  <svelte:fragment slot="title">Schedule report</svelte:fragment>
  <svelte:fragment slot="body">
    <BaseScheduledReportForm
      formId="create-scheduled-report-form"
      {formState}
    />
  </svelte:fragment>
  <svelte:fragment slot="footer">
    <div class="flex items-center gap-x-2 mt-2">
      {#if $createReport.isError}
        <div class="text-red-500">{$createReport.error.message}</div>
      {/if}
      <div class="grow" />
      <Button on:click={() => dispatch("close")} type="secondary">
        Cancel
      </Button>
      <Button
        disabled={$isSubmitting || $form["recipients"].length === 0}
        form="create-scheduled-report-form"
        submitForm
        type="primary"
      >
        Create
      </Button>
    </div>
  </svelte:fragment>
</Dialog>
