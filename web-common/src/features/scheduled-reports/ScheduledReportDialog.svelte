<script lang="ts" context="module">
  export type CreateReportProps = {
    mode: "create";
    query: V1Query;
    exploreName: string;
  };

  // Creates a canvas PDF report: no query, rendered from the canvas at download time.
  export type CreateCanvasReportProps = {
    mode: "create-canvas";
    canvasName: string;
  };

  export type EditReportProps = {
    mode: "edit";
    reportSpec: V1ReportSpec;
  };
</script>

<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
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
    getNewCanvasReportInitialFormValues,
    getNewReportInitialFormValues,
    getQueryNameFromQuery,
    parseMetricsViewFiltersAnnotation,
    stripInternalReportParams,
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
    V1ExportFormat,
    type V1MetricsViewAggregationRequest,
    type V1Query,
    type V1ReportSpec,
    type V1ReportSpecAnnotations,
  } from "../../runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { getCanvasFilters } from "../canvas/filters/canvas-filter-expressions";
  import { getCanvasStoreUnguarded } from "../canvas/state-managers/state-managers";
  import { getStateManagers } from "../dashboards/state-managers/state-managers";
  import { ResourceKind } from "../entity-management/resource-selectors";
  import BaseScheduledReportForm from "./BaseScheduledReportForm.svelte";
  import { convertFormValuesToCronExpression } from "./time-utils";

  export let open: boolean;
  export let props:
    | CreateReportProps
    | CreateCanvasReportProps
    | EditReportProps;

  const user = createAdminServiceGetCurrentUser();
  const FORM_ID = "scheduled-report-form";

  const runtimeClient = useRuntimeClient();

  // The dialog is mounted fresh for each report, so deriving these once at init is safe.
  const isCanvasReport =
    props.mode === "create-canvas" ||
    (props.mode === "edit" && !!props.reportSpec.annotations?.canvas);
  const canvasName =
    props.mode === "create-canvas"
      ? props.canvasName
      : props.mode === "edit"
        ? (props.reportSpec.annotations?.canvas ?? "")
        : "";

  $: ({ organization, project, report: reportName } = $page.params);
  $: ({ instanceId } = runtimeClient);

  $: listProjectMemberUsersQuery = createAdminServiceListProjectMemberUsers(
    organization,
    project,
  );
  $: projectMembersSet = new Set(
    $listProjectMemberUsersQuery.data?.members?.map((m) => m.userEmail) ?? [],
  );

  $: exploreName = isCanvasReport
    ? ""
    : props.mode === "create"
      ? props.exploreName
      : props.mode === "edit"
        ? getDashboardNameFromReport(props.reportSpec)
        : "";

  $: validExploreSpec = useExploreValidSpec(runtimeClient, exploreName);
  $: exploreSpec = $validExploreSpec.data?.explore ?? {};
  $: metricsViewName = exploreSpec.metricsView ?? "";

  $: allTimeRangeResp = useMetricsViewTimeRange(
    runtimeClient,
    metricsViewName,
    undefined,
    queryClient,
  );

  $: mutation =
    props.mode === "edit"
      ? createAdminServiceEditReport()
      : createAdminServiceCreateReport();

  $: queryName =
    props.mode === "create"
      ? getQueryNameFromQuery(props.query)
      : props.mode === "edit"
        ? ((props.reportSpec.resolverProperties?.query_name as
            | string
            | undefined) ?? props.reportSpec.queryName)
        : undefined;
  // Canvas reports have no query, and their queryArgsJson is an empty string (not undefined),
  // so they must not go through JSON.parse.
  $: aggregationRequest = (
    props.mode === "create"
      ? props.query.metricsViewAggregationRequest
      : props.mode === "edit" && !isCanvasReport
        ? JSON.parse(
            (props.reportSpec.resolverProperties?.query_args_json as
              | string
              | undefined) ||
              props.reportSpec.queryArgsJson ||
              "{}",
          )
        : {}
  ) as V1MetricsViewAggregationRequest;

  $: ({ filters, timeControls } = isCanvasReport
    ? { filters: undefined, timeControls: undefined }
    : getFiltersAndTimeControlsFromAggregationRequest(
        runtimeClient,
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

  // Canvas state is URL-param based. In create mode, the canvas page's current URL is the
  // state to display and save. In edit mode, the report's stored state is applied to the
  // form's canvas store directly (via urlStateOverride), never through the page URL: the
  // dialog is on the report page, whose URL must not be rewritten.
  const canvasStateOverride =
    props.mode === "edit" && isCanvasReport
      ? stripInternalReportParams(
          new URLSearchParams(props.reportSpec.annotations?.web_open_state),
        ).toString()
      : undefined;

  const schema = yup(
    object({
      title: string().required(m.report_form_required()),
      webOpenMode: string().required(m.report_form_required()),
      emailRecipients: array().of(
        string().email(m.report_form_invalid_email()),
      ),
      enableSlackNotification: boolean(), // Needed to get the type for validation
      slackChannels: array().of(string()),
      slackUsers: array().of(string().email(m.report_form_invalid_email())),
      columns: isCanvasReport
        ? array().of(string())
        : array().of(string()).min(1),
    })
      .test(
        "at-least-one-recipient",
        m.report_email_validation(),
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
        m.report_form_recipients_must_be_project(),
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
      : props.mode === "create-canvas"
        ? getNewCanvasReportInitialFormValues($user.data?.user?.email)
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

  let localError: string | undefined = undefined;
  $: generalErrors =
    $errors._errors?.[0] ?? localError ?? $mutation.error?.message;

  function buildReportOptions(values: ReportValues) {
    const refreshCron = convertFormValuesToCronExpression(
      values.frequency,
      values.dayOfWeek,
      values.timeOfDay,
      values.dayOfMonth,
    );
    const commonOptions = {
      displayName: values.title,
      refreshCron: refreshCron, // for testing: "* * * * *"
      refreshTimeZone: values.timeZone,
      emailRecipients: values.emailRecipients.filter(Boolean),
      slackChannels: values.enableSlackNotification
        ? values.slackChannels.filter(Boolean)
        : undefined,
      slackUsers: values.enableSlackNotification
        ? values.slackUsers.filter(Boolean)
        : undefined,
      webOpenMode: values.webOpenMode,
    };

    if (isCanvasReport) {
      // The canvas state (filters, time range, tabs) the report will render with,
      // plus the PDF rendering options, which the export page reads back from this state.
      // Create mode captures the canvas page's URL; edit mode reuses the report's stored
      // state (the dialog is on the report page, whose URL carries no canvas state).
      const stateParams = new URLSearchParams(
        props.mode === "edit"
          ? (props.reportSpec.annotations?.web_open_state ?? "")
          : window.location.search,
      );
      stateParams.set("pdf_include_filters", String(values.pdfIncludeFilters));
      stateParams.set("pdf_all_tabs", String(values.pdfAllTabs));

      // The selected filters are also baked into the report's security rules,
      // so magic-token recipients cannot query data beyond them.
      // The filter capture must not silently fail open: if the canvas store is
      // not available, fall back to the report's stored filters in edit mode,
      // and refuse to submit in create mode.
      const canvasStore = getCanvasStoreUnguarded(canvasName, instanceId);
      let metricsViewFilters = canvasStore
        ? getCanvasFilters(canvasStore.canvasEntity)
        : undefined;
      if (!canvasStore) {
        if (props.mode === "edit") {
          metricsViewFilters = parseMetricsViewFiltersAnnotation(
            props.reportSpec.annotations?.metrics_view_filters,
          );
        } else {
          throw new Error(m.report_form_canvas_loading());
        }
      }

      return {
        ...commonOptions,
        canvas: canvasName,
        exportFormat: V1ExportFormat.EXPORT_FORMAT_PDF,
        webOpenState: stateParams.toString(),
        metricsViewFilters,
      };
    }

    const filtersState = filters!.toState();
    const timeControlsState = timeControls!.toState();
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
    return {
      ...commonOptions,
      explore: exploreName,
      queryName: queryName,
      queryArgsJson: JSON.stringify(updatedAggregationRequest),
      exportLimit: values.exportLimit || undefined,
      exportIncludeHeader: values.exportIncludeHeader || false,
      exportFormat: values.exportFormat,
      webOpenState:
        props.mode === "create"
          ? currentProtobufState
          : props.mode === "edit"
            ? (props.reportSpec.annotations as V1ReportSpecAnnotations)[
                "web_open_state"
              ]
            : undefined,
    };
  }

  async function handleSubmit(values: ReportValues) {
    localError = undefined;
    let options: ReturnType<typeof buildReportOptions>;
    try {
      options = buildReportOptions(values);
    } catch (e) {
      localError = e instanceof Error ? e.message : String(e);
      return;
    }
    try {
      await $mutation.mutateAsync({
        org: organization,
        project,
        name: reportName,
        data: {
          options,
        },
      });

      if (props.mode === "edit") {
        await queryClient.invalidateQueries({
          queryKey: getRuntimeServiceGetResourceQueryKey(instanceId, {
            name: {
              name: reportName,
              kind: ResourceKind.Report,
            },
          }),
        });
      }

      await queryClient.invalidateQueries({
        queryKey: getRuntimeServiceListResourcesQueryKey(instanceId),
      });

      open = false;

      eventBus.emit("notification", {
        message:
          props.mode === "edit"
            ? m.report_form_edited_notification()
            : m.report_form_created_notification(),
        link:
          props.mode === "edit"
            ? undefined
            : {
                href: `/${organization}/${project}/-/reports`,
                text: m.report_form_go_to_reports(),
              },
        type: "success",
      });
    } catch {
      // showing error below
    }
  }
</script>

<Dialog.Root bind:open>
  <Dialog.Content class="min-w-[900px]" escapeKeydownBehavior="ignore">
    <Dialog.Title>{m.report_form_schedule()}</Dialog.Title>

    <BaseScheduledReportForm
      formId={FORM_ID}
      data={form}
      {errors}
      {submit}
      {enhance}
      exploreName={exploreName ?? ""}
      {canvasName}
      {canvasStateOverride}
      {filters}
      {timeControls}
    />

    {#if generalErrors}
      <div class="text-red-500">{generalErrors}</div>
    {/if}
    <div class="flex items-center gap-x-2 mt-5">
      <div class="grow"></div>
      <Button onClick={() => (open = false)} type="secondary"
        >{m.report_form_cancel()}</Button
      >
      <Button
        disabled={$submitting}
        form={FORM_ID}
        submitForm
        type="primary"
        label={props.mode === "edit"
          ? m.report_form_save()
          : m.report_form_create()}
      >
        {props.mode === "edit"
          ? m.report_form_save_button()
          : m.report_form_create_button()}
      </Button>
    </div>
  </Dialog.Content>
</Dialog.Root>
