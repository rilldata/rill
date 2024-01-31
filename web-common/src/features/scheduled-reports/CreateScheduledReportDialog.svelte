<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceCreateReport,
    createAdminServiceGetCurrentUser,
  } from "@rilldata/web-admin/client";
  import Dialog from "@rilldata/web-common/components/dialog-v2/Dialog.svelte";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { createEventDispatcher } from "svelte";
  import { createForm } from "svelte-forms-lib";
  import * as yup from "yup";
  import { Button } from "../../components/button";
  import { notifications } from "../../components/notifications";
  import { getLocalIANA } from "../../lib/time/timezone";
  import {
    V1ExportFormat,
    getRuntimeServiceListResourcesQueryKey,
  } from "../../runtime-client";
  import { runtime } from "../../runtime-client/runtime-store";
  import BaseScheduledReportForm from "./BaseScheduledReportForm.svelte";
  import {
    convertFormValuesToCronExpression,
    getNextQuarterHour,
    getTimeIn24FormatFromDateTime,
    getTodaysDayOfWeek,
  } from "./time-utils";

  export let open: boolean;
  export let queryName: string;
  export let queryArgs: any;

  const user = createAdminServiceGetCurrentUser();
  const createReport = createAdminServiceCreateReport();
  $: organization = $page.params.organization;
  $: project = $page.params.project;
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
      recipients: [
        { email: $user.data?.user?.email ? $user.data.user.email : "" },
        { email: "" },
      ],
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
      const refreshCron = convertFormValuesToCronExpression(
        values.frequency,
        values.dayOfWeek,
        values.timeOfDay,
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
              recipients: values.recipients.map((r) => r.email).filter(Boolean),
            },
          },
        });
        queryClient.invalidateQueries(
          getRuntimeServiceListResourcesQueryKey($runtime.instanceId),
        );
        dispatch("close");
        notifications.send({
          message: "Report created",
          link: {
            href: `/${organization}/${project}/-/reports`,
            text: "Go to scheduled reports",
          },
          options: {
            persistedLink: true,
          },
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
      metricsViewName={queryArgs.metricsViewName}
    />
  </svelte:fragment>
  <svelte:fragment slot="footer">
    <div class="flex items-center gap-x-2 mt-5">
      {#if $createReport.isError}
        <div class="text-red-500">{$createReport.error.message}</div>
      {/if}
      <div class="grow" />
      <Button on:click={() => dispatch("close")} type="secondary">
        Cancel
      </Button>
      <Button
        disabled={$isSubmitting ||
          $form["recipients"].filter((r) => r.email).length === 0}
        form="create-scheduled-report-form"
        submitForm
        type="primary"
      >
        Create
      </Button>
    </div>
  </svelte:fragment>
</Dialog>
