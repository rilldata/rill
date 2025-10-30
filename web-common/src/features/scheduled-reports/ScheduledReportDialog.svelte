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
    createAdminServiceListProjectMemberUsers,
  } from "@rilldata/web-admin/client";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import {
    aggregationRequestWithFilters,
    aggregationRequestWithRowsAndColumns,
    aggregationRequestWithTimeRange,
    buildAggregationRequest,
  } from "@rilldata/web-common/features/dashboards/aggregation-request-utils.ts";
  import { useMetricsViewTimeRange } from "@rilldata/web-common/features/dashboards/selectors.ts";
  import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors.ts";
  import {
    getDashboardNameFromReport,
    getExistingReportInitialFormValues,
    getFiltersAndTimeControlsFromAggregationRequest,
    getNewReportInitialFormValues,
    getQueryNameFromQuery,
    ReportRunAs,
    type ReportValues,
  } from "@rilldata/web-common/features/scheduled-reports/utils";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { get } from "svelte/store";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup, type ValidationAdapter } from "sveltekit-superforms/adapters";
  import { array, object, string, boolean } from "yup";
  import { Button } from "../../components/button";
  import {
    getRuntimeServiceGetResourceQueryKey,
    getRuntimeServiceListResourcesQueryKey,
    type V1MetricsViewAggregationRequest,
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
  const FORM_ID = "scheduled-report-form";

  $: ({ organization, project, report: reportName } = $page.params);
  $: ({ instanceId } = $runtime);

  $: listProjectMemberUsersQuery = createAdminServiceListProjectMemberUsers(
    organization,
    project,
  );
  $: projectMembersSet = new Set(
    $listProjectMemberUsersQuery.data?.members?.map((m) => m.userEmail) ?? [],
  );

  $: exploreName =
    props.mode === "create"
      ? props.exploreName
      : getDashboardNameFromReport(props.reportSpec);

  $: validExploreSpec = useExploreValidSpec(instanceId, exploreName);
  $: exploreSpec = $validExploreSpec.data?.explore ?? {};
  $: metricsViewName = exploreSpec.metricsView ?? "";

  $: allTimeRangeResp = useMetricsViewTimeRange(
    instanceId,
    metricsViewName,
    undefined,
    queryClient,
  );

  $: mutation =
    props.mode === "create"
      ? createAdminServiceCreateReport()
      : createAdminServiceEditReport();

  $: queryName =
    props.mode === "create"
      ? getQueryNameFromQuery(props.query)
      : props.reportSpec.queryName;
  $: aggregationRequest = (
    props.mode === "create"
      ? props.query.metricsViewAggregationRequest
      : JSON.parse(props.reportSpec.queryArgsJson || "{}")
  ) as V1MetricsViewAggregationRequest;

  $: ({ filters, timeControls } =
    getFiltersAndTimeControlsFromAggregationRequest(
      instanceId,
      metricsViewName,
      exploreName,
      aggregationRequest,
      $allTimeRangeResp.data?.timeRangeSummary,
    ));

  let currentProtobufState: string | undefined = undefined;
  if (open && props.mode === "create") {
    const stateManagers = getStateManagers();
    const { dashboardStore } = stateManagers;
    currentProtobufState = get(dashboardStore).proto;
  }

  const schema = yup(
    object({
      title: string().required("Required"),
      webOpenMode: string().required("Required"),
      emailRecipients: array().of(string().email("Invalid email")),
      enableSlackNotification: boolean(), // Needed to get the type for validation
      slackChannels: array().of(string()),
      slackUsers: array().of(string().email("Invalid email")),
      columns: array().of(string()).min(1),
    })
      .test(
        "at-least-one-recipient",
        "At least one email recipient, slack user, or slack channel is required",
        function (value) {
          // Check if at least one array has non-empty values
          const hasEmailRecipients = value.emailRecipients
            ? value.emailRecipients.filter(Boolean).length > 0
            : false;
          if (!value.enableSlackNotification) return hasEmailRecipients;

          const hasSlackUsers = value.slackUsers
            ? value.slackUsers.filter(Boolean).length > 0
            : false;
          const hasSlackChannels = value.slackChannels
            ? value.slackChannels.filter(Boolean).length > 0
            : false;

          return hasEmailRecipients || hasSlackUsers || hasSlackChannels;
        },
      )
      .test(
        "as-recipients-in-project",
        "Recipients must be part of the project when running as recipient",
        function (values) {
          if (values.webOpenMode !== ReportRunAs.Recipient) return true;

          return (
            values.emailRecipients?.every(
              (recipient) => !recipient || projectMembersSet.has(recipient),
            ) ?? true
          );
        },
      ),
  ) as ValidationAdapter<ReportValues>;

  $: initialValues =
    props.mode === "create"
      ? getNewReportInitialFormValues(
          $user.data?.user?.email,
          aggregationRequest,
        )
      : getExistingReportInitialFormValues(
          props.reportSpec,
          $user.data?.user?.email,
          aggregationRequest,
        );

  $: ({ form, errors, enhance, submit, submitting } = superForm(
    defaults(initialValues, schema),
    {
      id: FORM_ID,
      SPA: true,
      validators: schema,
      async onUpdate({ form }) {
        if (!form.valid) return;
        const values = form.data;
        return handleSubmit(values);
      },
      // We need to run the 1st validation only after a submit.
      // But successive validations should be on input.
      // Here, "auto" achieves this.
      validationMethod: "auto",
      invalidateAll: false,
    },
  ));

  $: generalErrors = $errors._errors?.[0] ?? $mutation.error?.message;

  async function handleSubmit(values: ReportValues) {
    const refreshCron = convertFormValuesToCronExpression(
      values.frequency,
      values.dayOfWeek,
      values.timeOfDay,
      values.dayOfMonth,
    );
    const filtersState = filters.toState();
    const timeControlsState = timeControls.toState();
    const updatedAggregationRequest = buildAggregationRequest(
      aggregationRequest,
      [
        aggregationRequestWithTimeRange(exploreSpec, timeControlsState),
        aggregationRequestWithFilters(filtersState),
        aggregationRequestWithRowsAndColumns({
          exploreSpec,
          rows: values.rows,
          columns: values.columns,
          showTimeComparison: timeControlsState.showTimeComparison,
          selectedTimezone: timeControlsState.selectedTimezone,
        }),
      ],
    );

    try {
      await $mutation.mutateAsync({
        org: organization,
        project,
        name: reportName,
        data: {
          options: {
            displayName: values.title,
            refreshCron: refreshCron, // for testing: "* * * * *"
            refreshTimeZone: values.timeZone,
            explore: exploreName,
            queryName: queryName,
            queryArgsJson: JSON.stringify(updatedAggregationRequest),
            exportLimit: values.exportLimit || undefined,
            exportIncludeHeader: values.exportIncludeHeader || false,
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
            webOpenMode: values.webOpenMode,
          },
        },
      });

      if (props.mode === "edit") {
        await queryClient.invalidateQueries({
          queryKey: getRuntimeServiceGetResourceQueryKey(instanceId, {
            "name.name": reportName,
            "name.kind": ResourceKind.Report,
          }),
        });
      }

      await queryClient.invalidateQueries({
        queryKey: getRuntimeServiceListResourcesQueryKey(instanceId),
      });

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

<Dialog.Root bind:open closeOnEscape={false}>
  <Dialog.Content class="min-w-[900px]">
    <Dialog.Title>Schedule report</Dialog.Title>

    <BaseScheduledReportForm
      formId={FORM_ID}
      data={form}
      {errors}
      {submit}
      {enhance}
      exploreName={exploreName ?? ""}
      {filters}
      {timeControls}
    />

    {#if generalErrors}
      <div class="text-red-500">{generalErrors}</div>
    {/if}
    <div class="flex items-center gap-x-2 mt-5">
      <div class="grow" />
      <Button onClick={() => (open = false)} type="secondary">Cancel</Button>
      <Button
        disabled={$submitting}
        form={FORM_ID}
        submitForm
        type="primary"
        label={props.mode === "create" ? "Create report" : "Save report"}
      >
        {props.mode === "create" ? "Create" : "Save"}
      </Button>
    </div>
  </Dialog.Content>
</Dialog.Root>
