<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceCreateReport,
    createAdminServiceEditReport,
    createAdminServiceGetCurrentUser,
    ReportOptionsOpenMode,
  } from "@rilldata/web-admin/client";
  import * as Dialog from "@rilldata/web-common/components/dialog-v2";
  import {
    getDashboardNameFromReport,
    getInitialValues,
    getQueryArgsFromQuery,
    getQueryNameFromQuery,
    type ReportValues,
  } from "@rilldata/web-common/features/scheduled-reports/utils";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { get } from "svelte/store";
  import { defaults, superForm } from "sveltekit-superforms";
  import { type ValidationAdapter, yup } from "sveltekit-superforms/adapters";
  import { array, object, string } from "yup";
  import { Button } from "../../components/button";
  import {
    getRuntimeServiceGetResourceQueryKey,
    getRuntimeServiceListResourcesQueryKey,
    type V1Query,
    type V1ReportSpec,
    type V1ReportSpecAnnotations,
  } from "../../runtime-client";
  import { runtime } from "../../runtime-client/runtime-store";
  import { getStateManagers } from "../dashboards/state-managers/state-managers";
  import { ResourceKind } from "../entity-management/resource-selectors";
  import BaseScheduledReportForm from "./BaseScheduledReportForm.svelte";
  import { convertFormValuesToCronExpression } from "./time-utils";

  export let open: boolean;
  export let query: V1Query | undefined = undefined;
  export let exploreName: string | undefined = undefined;
  export let reportSpec: V1ReportSpec | undefined = undefined;

  $: ({ instanceId } = $runtime);

  $: isEdit = !!reportSpec;

  const user = createAdminServiceGetCurrentUser();

  $: if (!exploreName) {
    exploreName = getDashboardNameFromReport(reportSpec) ?? "";
  }

  $: queryName = query ? getQueryNameFromQuery(query) : undefined;
  $: queryArgs = query ? getQueryArgsFromQuery(query) : undefined;

  let currentProtobufState: string | undefined = undefined;
  if (open && !isEdit) {
    const stateManagers = getStateManagers();
    const { dashboardStore } = stateManagers;
    currentProtobufState = get(dashboardStore).proto;
  }

  $: ({ organization, project, report: reportName } = $page.params);

  $: mutation = isEdit
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
  ) as ValidationAdapter<ReportValues>;

  async function handleSubmit(values: ReportValues) {
    const refreshCron = convertFormValuesToCronExpression(
      values.frequency,
      values.dayOfWeek,
      values.timeOfDay,
      values.dayOfMonth,
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
            explore: exploreName,
            queryName: reportSpec?.queryName ?? queryName,
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
              : currentProtobufState,
            webOpenMode: isEdit
              ? (((reportSpec?.annotations as V1ReportSpecAnnotations)[
                  "web_open_mode"
                ] as ReportOptionsOpenMode) ??
                ReportOptionsOpenMode.OPEN_MODE_RECIPIENT) // Backwards compatibility
              : ReportOptionsOpenMode.OPEN_MODE_CREATOR,
          },
        },
      });

      if (isEdit) {
        await queryClient.invalidateQueries(
          getRuntimeServiceGetResourceQueryKey(instanceId, {
            "name.name": reportName,
            "name.kind": ResourceKind.Report,
          }),
        );
      }

      await queryClient.invalidateQueries(
        getRuntimeServiceListResourcesQueryKey(instanceId),
      );

      open = false;

      eventBus.emit("notification", {
        message: `Report ${isEdit ? "edited" : "created"}`,
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
      invalidateAll: false,
    },
  );
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
        {isEdit ? "Save" : "Create"}
      </Button>
    </div>
  </Dialog.Content>
</Dialog.Root>
