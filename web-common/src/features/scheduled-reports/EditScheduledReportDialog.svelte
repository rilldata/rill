<script lang="ts">
  import { page } from "$app/stores";
  import { createAdminServiceEditReport } from "@rilldata/web-admin/client";
  import Dialog from "@rilldata/web-common/components/dialog/Dialog.svelte";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { createEventDispatcher } from "svelte";
  import { createForm } from "svelte-forms-lib";
  import * as yup from "yup";
  import { Button } from "../../components/button";
  import { notifications } from "../../components/notifications";
  import {
    getRuntimeServiceGetResourceQueryKey,
    getRuntimeServiceListResourcesQueryKey,
    V1ExportFormat,
    V1ReportSpec,
    V1ReportSpecAnnotations,
  } from "../../runtime-client";
  import { runtime } from "../../runtime-client/runtime-store";
  import { ResourceKind } from "../entity-management/resource-selectors";
  import BaseScheduledReportForm from "./BaseScheduledReportForm.svelte";
  import {
    convertFormValuesToCronExpression,
    getDayOfWeekFromCronExpression,
    getFrequencyFromCronExpression,
    getTimeOfDayFromCronExpression,
  } from "./time-utils";

  export let open: boolean;
  export let reportSpec: V1ReportSpec;

  const editReport = createAdminServiceEditReport();
  $: organization = $page.params.organization;
  $: project = $page.params.project;
  $: reportName = $page.params.report;
  $: metricsViewName = reportSpec?.queryArgsJson
    ? JSON.parse(reportSpec.queryArgsJson).metricsViewName
    : "";

  const queryClient = useQueryClient();
  const dispatch = createEventDispatcher();

  const formState = createForm({
    initialValues: {
      title: reportSpec.title as string,
      frequency: getFrequencyFromCronExpression(
        reportSpec.refreshSchedule?.cron as string,
      ),
      dayOfWeek: getDayOfWeekFromCronExpression(
        reportSpec.refreshSchedule?.cron as string,
      ),
      timeOfDay: getTimeOfDayFromCronExpression(
        reportSpec.refreshSchedule?.cron as string,
      ),
      timeZone: reportSpec.refreshSchedule?.timeZone as string, // all UI-created reports have a timeZone
      exportFormat:
        reportSpec.exportFormat ?? V1ExportFormat.EXPORT_FORMAT_UNSPECIFIED,
      exportLimit: reportSpec.exportLimit === "0" ? "" : reportSpec.exportLimit,
      recipients:
        reportSpec.notifySpec?.notifiers
          ?.find((n) => n.connector === "email")
          ?.email?.recipients?.map((email) => ({
            email: email,
          })) ?? [],
    },
    validationSchema: yup.object({
      title: yup.string().required("Required"),
      recipients: yup.array().of(
        yup.object().shape({
          email: yup.string().email("Invalid email"),
        }),
      ),
    }),
    onSubmit: async (values) => {
      const queryName = reportSpec.queryName ?? "";
      const queryArgs = reportSpec.queryArgsJson
        ? JSON.parse(reportSpec.queryArgsJson)
        : {};
      const refreshCron = convertFormValuesToCronExpression(
        values.frequency,
        values.dayOfWeek,
        values.timeOfDay,
      );

      try {
        await $editReport.mutateAsync({
          organization,
          project,
          name: reportName,
          data: {
            options: {
              title: values.title,
              refreshCron: refreshCron,
              refreshTimeZone: values.timeZone,
              queryName: queryName,
              queryArgsJson: JSON.stringify(queryArgs),
              exportLimit: values.exportLimit || undefined,
              exportFormat: values.exportFormat,
              openProjectSubpath: (
                reportSpec.annotations as V1ReportSpecAnnotations
              )["web_open_project_subpath"],
              emailRecipients: values.recipients
                .map((r) => r.email)
                .filter(Boolean),
            },
          },
        });
        queryClient.invalidateQueries(
          getRuntimeServiceGetResourceQueryKey($runtime.instanceId, {
            "name.name": reportName,
            "name.kind": ResourceKind.Report,
          }),
        );
        queryClient.invalidateQueries(
          getRuntimeServiceListResourcesQueryKey($runtime.instanceId),
        );
        dispatch("close");
        notifications.send({
          message: "Report edited",
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
  <svelte:fragment slot="title">Edit scheduled report</svelte:fragment>
  <svelte:fragment slot="body">
    <BaseScheduledReportForm
      formId="edit-scheduled-report-form"
      {formState}
      {metricsViewName}
    />
  </svelte:fragment>
  <svelte:fragment slot="footer">
    <div class="flex items-center gap-x-2 mt-2">
      {#if $editReport.isError}
        <div class="text-red-500">{$editReport.error.message}</div>
      {/if}
      <div class="grow" />
      <Button on:click={() => dispatch("close")} type="secondary">
        Cancel
      </Button>
      <Button
        disabled={$isSubmitting || $form["recipients"].length === 0}
        form="edit-scheduled-report-form"
        submitForm
        type="primary"
      >
        Save
      </Button>
    </div>
  </svelte:fragment>
</Dialog>
