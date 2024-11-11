<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceCreateReport,
    createAdminServiceGetCurrentUser,
    createAdminServiceEditReport,
  } from "@rilldata/web-admin/client";
  import {
    getDashboardNameFromReport,
    getInitialValues,
  } from "@rilldata/web-common/features/scheduled-reports/utils";
  import { createForm } from "svelte-forms-lib";
  import { defaults, superForm } from "sveltekit-superforms";
  import { array, object, string } from "yup";
  import { yup } from "sveltekit-superforms/adapters";
  import { Button } from "../../components/button";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import {
    getRuntimeServiceListResourcesQueryKey,
    type V1ReportSpec,
    getRuntimeServiceGetResourceQueryKey,
    type V1ReportSpecAnnotations,
  } from "../../runtime-client";
  import { runtime } from "../../runtime-client/runtime-store";
  import BaseScheduledReportForm from "./BaseScheduledReportForm.svelte";
  import { convertFormValuesToCronExpression } from "./time-utils";
  import * as Dialog from "@rilldata/web-common/components/dialog-v2";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { ResourceKind } from "../entity-management/resource-selectors";

  export let open: boolean;
  export let queryArgs: any | undefined = undefined;
  export let metricsViewProto: string | undefined = undefined;
  export let exploreName: string | undefined = undefined;
  export let reportSpec: V1ReportSpec | undefined = undefined;

  const user = createAdminServiceGetCurrentUser();

  $: if (!exploreName) {
    exploreName =
      getDashboardNameFromReport(reportSpec) ?? queryArgs.metricsViewName;
  }

  $: ({ organization, project, report: reportName } = $page.params);

  $: mutation = reportSpec
    ? createAdminServiceEditReport()
    : createAdminServiceCreateReport();

  const initialValues = getInitialValues(reportSpec, $user.data?.user?.email);
  const schema = yup(
    object({
      title: string().required("Required"),
      emailRecipients: array().of(string().email("Invalid email")),
      slackChannels: array().of(string()),
      slackUsers: array().of(string().email("Invalid email")),
    }),
  );

  async function handleSubmit(values: ReturnType<typeof getInitialValues>) {
    const refreshCron = convertFormValuesToCronExpression(
      values.frequency,
      values.dayOfWeek,
      values.timeOfDay,
    );

    try {
      await $mutation.mutateAsync({
        organization,
        project,
        name: reportName,
        data: {
          options: {
            displayName: values.title,
            refreshCron: refreshCron, // for testing: "* * * * *"
            refreshTimeZone: values.timeZone,
            queryName: reportSpec?.queryName ?? "MetricsViewAggregation",
            queryArgsJson: JSON.stringify(
              reportSpec?.queryArgsJson
                ? JSON.parse(reportSpec.queryArgsJson)
                : queryArgs,
            ),
            exportLimit: values.exportLimit || undefined,
            exportFormat: values.exportFormat,
            emailRecipients: values.emailRecipients.filter(Boolean),
            slackChannels: values.enableSlackNotification
              ? values.slackChannels.filter(Boolean)
              : undefined,
            slackUsers: values.enableSlackNotification
              ? values.slackUsers.filter(Boolean)
              : undefined,
            webOpenState: reportSpec
              ? (reportSpec.annotations as V1ReportSpecAnnotations)[
                  "web_open_state"
                ]
              : metricsViewProto,
            webOpenPath: exploreName ? `/explore/${exploreName}` : undefined,
          },
        },
      });

      if (reportSpec) {
        await queryClient.invalidateQueries(
          getRuntimeServiceGetResourceQueryKey($runtime.instanceId, {
            "name.name": reportName,
            "name.kind": ResourceKind.Report,
          }),
        );
      }

      await queryClient.invalidateQueries(
        getRuntimeServiceListResourcesQueryKey($runtime.instanceId),
      );

      open = false;

      eventBus.emit("notification", {
        message: `Report ${reportSpec ? "edited" : "created"}`,
        link: reportSpec
          ? undefined
          : {
              href: `/${organization}/${project}/-/reports`,
              text: "Go to scheduled reports",
            },
        type: "success",
      });
    } catch {
      // showing error below
    }
  }

  const { form, errors, enhance, submit, submitting } = superForm(
    defaults(initialValues, schema),
    {
      SPA: true,
      validators: schema,
      async onUpdate({ form }) {
        if (!form.valid) return;
        const values = form.data;
        return handleSubmit(values);
      },
      validationMethod: "oninput",
    },
  );
  $: console.log($form, $errors);
</script>

<Dialog.Root bind:open>
  <Dialog.Content>
    <Dialog.Title>Schedule report</Dialog.Title>

    <BaseScheduledReportForm
      formId="scheduled-report-form"
      data={form}
      {errors}
      {submit}
      {enhance}
      exploreName={exploreName ?? ""}
    />

    <div class="flex items-center gap-x-2 mt-5">
      {#if $mutation.isError}
        <div class="text-red-500">{$mutation.error.message}</div>
      {/if}
      <div class="grow" />
      <Button on:click={() => (open = false)} type="secondary">Cancel</Button>
      <Button
        disabled={$submitting || $form["emailRecipients"]?.length === 0}
        form="scheduled-report-form"
        submitForm
        type="primary"
      >
        {reportSpec ? "Save" : "Create"}
      </Button>
    </div>
  </Dialog.Content>
</Dialog.Root>
