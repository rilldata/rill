<script lang="ts" context="module">
  export type CreateReportProps = {
    mode: "create";
    query: V1Query;
    exploreName: string;
  };

  export type EditReportProps = {
    mode: "edit";
    reportSpec: V1ReportSpec;
  };
</script>

<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceCreateReport,
    createAdminServiceEditReport,
    createAdminServiceGetCurrentUser,
  } from "@rilldata/web-admin/client";
  import * as Dialog from "@rilldata/web-common/components/dialog-v2";
  import {
    getDashboardNameFromReport,
    getExistingReportInitialFormValues,
    getNewReportInitialFormValues,
    getQueryArgsJsonFromQuery,
    getQueryNameFromQuery,
    type ReportValues,
  } from "@rilldata/web-common/features/scheduled-reports/utils";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { get } from "svelte/store";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup, type ValidationAdapter } from "sveltekit-superforms/adapters";
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
  export let props: CreateReportProps | EditReportProps;

  const user = createAdminServiceGetCurrentUser();

  $: ({ organization, project, report: reportName } = $page.params);
  $: ({ instanceId } = $runtime);

  $: mutation =
    props.mode === "create"
      ? createAdminServiceCreateReport()
      : createAdminServiceEditReport();

  $: queryName =
    props.mode === "create"
      ? getQueryNameFromQuery(props.query)
      : props.reportSpec.queryName;
  $: queryArgsJson =
    props.mode === "create"
      ? getQueryArgsJsonFromQuery(props.query)
      : props.reportSpec.queryArgsJson;

  $: exploreName =
    props.mode === "create"
      ? props.exploreName
      : getDashboardNameFromReport(props.reportSpec);

  let currentProtobufState: string | undefined = undefined;
  if (open && props.mode === "create") {
    const stateManagers = getStateManagers();
    const { dashboardStore } = stateManagers;
    currentProtobufState = get(dashboardStore).proto;
  }

  const schema = yup(
    object({
      title: string().required("Required"),
      emailRecipients: array().of(string().email("Invalid email")),
      slackChannels: array().of(string()),
      slackUsers: array().of(string().email("Invalid email")),
    }),
  ) as ValidationAdapter<ReportValues>;

  $: initialValues =
    props.mode === "create"
      ? getNewReportInitialFormValues($user.data?.user?.email)
      : getExistingReportInitialFormValues(
          props.reportSpec,
          $user.data?.user?.email,
        );

  $: ({ form, errors, enhance, submit, submitting } = superForm(
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
  ));

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
            queryName: queryName,
            queryArgsJson: queryArgsJson,
            exportLimit: values.exportLimit || undefined,
            exportFormat: values.exportFormat,
            emailRecipients: values.emailRecipients.filter(Boolean),
            slackChannels: values.enableSlackNotification
              ? values.slackChannels.filter(Boolean)
              : undefined,
            slackUsers: values.enableSlackNotification
              ? values.slackUsers.filter(Boolean)
              : undefined,
            webOpenState:
              props.mode === "create"
                ? currentProtobufState
                : (props.reportSpec.annotations as V1ReportSpecAnnotations)[
                    "web_open_state"
                  ],
            webOpenMode:
              props.mode === "create"
                ? "recipient" // To be changed to "filtered" once support is added
                : ((props.reportSpec.annotations as V1ReportSpecAnnotations)[
                    "web_open_mode"
                  ] ?? "recipient"), // Backwards compatibility
          },
        },
      });

      if (props.mode === "edit") {
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
        message: `Report ${props.mode === "create" ? "created" : "edited"}`,
        link:
          props.mode === "create"
            ? {
                href: `/${organization}/${project}/-/reports`,
                text: "Go to scheduled reports",
              }
            : undefined,
        type: "success",
      });
    } catch {
      // showing error below
    }
  }
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
        {props.mode === "create" ? "Create" : "Save"}
      </Button>
    </div>
  </Dialog.Content>
</Dialog.Root>
