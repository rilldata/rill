<script lang="ts" module>
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
  import { ephemeralDefsFromRequestMeasures } from "@rilldata/web-common/features/dashboards/ephemeral-measures/measure-mapping";
  import { useExploreState } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { page } from "$app/state";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import {
    aggregationRequestWithRowsAndColumns,
    aggregationRequestWithTimeRange,
    buildAggregationRequest,
  } from "@rilldata/web-common/features/dashboards/aggregation-request-utils.ts";
  import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors.ts";
  import { getSpecQuery } from "@rilldata/web-common/features/explore-mappers/utils";
  import {
    getDashboardNameFromReport,
    applyReportFiltersToParams,
    createReportFilterManagers,
    getExistingReportInitialFormValues,
    getFilterParamsFromAggregationRequest,
    getNewCanvasReportInitialFormValues,
    getNewReportInitialFormValues,
    getQueryNameFromQuery,
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
  import { getExportLimitSchema } from "@rilldata/web-common/features/exports/export-file-options";
  import { Button } from "../../components/button";
  import {
    createRuntimeServiceGetInstance,
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
  import {
    CanvasDashboardConfigProvider,
    ExploreDashboardConfigProvider,
  } from "@rilldata/web-common/features/dashboards/providers/DashboardConfigProvider.svelte.ts";
  import { untrack } from "svelte";
  import { createFormMetadataProvider } from "@rilldata/web-common/features/scheduled-reports/FormMetadataProvider.svelte.ts";
  import {
    reportDeleteMutationFactory,
    getReportMutationFactory,
  } from "@rilldata/web-common/features/scheduled-reports/ReportFormMetadataProvider.ts";
  import { EmbedStore } from "@rilldata/web-common/features/embeds/embed-store.ts";

  let {
    open = $bindable(),
    props,
  }: {
    open: boolean;
    props: CreateReportProps | CreateCanvasReportProps | EditReportProps;
  } = $props();

  const FORM_ID = "scheduled-report-form";

  const runtimeClient = useRuntimeClient();
  // svelte-ignore state_referenced_locally
  const provider = createFormMetadataProvider(
    runtimeClient,
    getReportMutationFactory(props.mode === "edit"),
    reportDeleteMutationFactory,
  );
  const { mutation } = provider;

  // The dialog is mounted fresh for each report, so deriving these once at init is safe.
  // svelte-ignore state_referenced_locally
  const isCanvasReport =
    props.mode === "create-canvas" ||
    (props.mode === "edit" && !!props.reportSpec.annotations?.canvas);
  // svelte-ignore state_referenced_locally
  const canvasName =
    props.mode === "create-canvas"
      ? props.canvasName
      : props.mode === "edit"
        ? (props.reportSpec.annotations?.canvas ?? "")
        : "";

  const { instanceId } = runtimeClient;
  const reportName = $derived(page.params.report);

  // The creator of a report created from an embed has a field for their email, which always gets the report:
  // in an embed, if rill.yaml configures `embeds.open_url`, since only that email gets the open link to the embedder's app;
  // and for reports with a saved creator email, which the admin service always keeps as a recipient.
  // Outside of an embed, the field is read-only, since only the creator can change it.
  const isEmbedded = EmbedStore.isEmbedded();
  const embedOpenUrlQuery = createRuntimeServiceGetInstance(
    runtimeClient,
    {},
    {
      query: {
        enabled: isEmbedded,
        select: (data) => data?.instance?.embedOpenUrl ?? "",
      },
    },
  );
  // svelte-ignore state_referenced_locally
  const hasSavedEmbedOwnerEmail =
    props.mode === "edit" &&
    !!props.reportSpec.annotations?.admin_embed_owner_email;
  const showEmbedOwnerEmail = $derived(
    hasSavedEmbedOwnerEmail || (isEmbedded && !!$embedOpenUrlQuery.data),
  );

  // svelte-ignore state_referenced_locally
  const exploreName = isCanvasReport
    ? ""
    : props.mode === "create"
      ? props.exploreName
      : props.mode === "edit"
        ? getDashboardNameFromReport(props.reportSpec)
        : "";

  const validExploreSpec = $derived(
    useExploreValidSpec(runtimeClient, exploreName),
  );
  const exploreSpec = $derived($validExploreSpec.data?.explore ?? {});
  const exploreStateStore = $derived(useExploreState(exploreName));

  const queryName = $derived(
    props.mode === "create"
      ? getQueryNameFromQuery(props.query)
      : props.mode === "edit"
        ? getSpecQuery(props.reportSpec).queryName
        : undefined,
  );
  // Canvas reports have no query, and their queryArgsJson is an empty string (not undefined),
  // so they must not go through JSON.parse.
  const aggregationRequest = $derived(
    (props.mode === "create"
      ? props.query.metricsViewAggregationRequest
      : props.mode === "edit" && !isCanvasReport
        ? JSON.parse(getSpecQuery(props.reportSpec).queryArgsJson || "{}")
        : {}) as V1MetricsViewAggregationRequest,
  );

  // The explore state only exists on the explore page (create mode);
  // an edited report carries its definitions in the saved request.
  const ephemeralMeasures = $derived(
    props.mode === "edit"
      ? ephemeralDefsFromRequestMeasures(aggregationRequest.measures)
      : $exploreStateStore?.ephemeralMeasures,
  );

  // Canvas state is URL-param based. In create mode, the canvas page's current URL is the
  // state to save. In edit mode, it is the report's stored state: the dialog is on the
  // report page, whose URL carries no canvas state.
  // svelte-ignore state_referenced_locally
  const canvasBaseParams = isCanvasReport
    ? stripInternalReportParams(
        new URLSearchParams(
          props.mode === "edit"
            ? props.reportSpec.annotations?.web_open_state
            : window.location.search,
        ),
      )
    : undefined;

  // Filters are edited in memory and only saved with the report, so the dashboard is left untouched.
  const dashboardConfigProvider = isCanvasReport
    ? new CanvasDashboardConfigProvider(runtimeClient, canvasName)
    : new ExploreDashboardConfigProvider(runtimeClient, exploreName);
  const {
    expressionFilterManager,
    timeFilterManager,
    cleanup: cleanupFilterManagers,
  } = createReportFilterManagers(
    runtimeClient,
    dashboardConfigProvider,
    // Explore urls carry the singular `f` param, canvas urls one per metrics view.
    !isCanvasReport,
    untrack(getInitialFilterParams),
  );
  $effect(() => () => {
    cleanupFilterManagers();
    dashboardConfigProvider.cleanup?.();
    provider.cleanup();
  });

  function getInitialFilterParams() {
    if (!canvasBaseParams) {
      return getFilterParamsFromAggregationRequest(aggregationRequest);
    }
    if (props.mode === "edit") return canvasBaseParams;

    // The canvas page's filter managers already have the defaults applied for an empty URL.
    const canvasStore = getCanvasStoreUnguarded(canvasName, instanceId);
    if (!canvasStore) return canvasBaseParams;
    const params = new URLSearchParams();
    canvasStore.canvasEntity.expressionFilterManager.applyFilterToParams(
      params,
    );
    canvasStore.canvasEntity.timeFilterManager.applyFilterToParams(params);
    return params;
  }

  let currentProtobufState: string | undefined = undefined;
  // svelte-ignore state_referenced_locally
  if (open && props.mode === "create") {
    const stateManagers = getStateManagers();
    const { dashboardStore } = stateManagers;
    currentProtobufState = get(dashboardStore).proto;
  }

  const schema = yup(
    object({
      title: string().required(m.report_form_required()),
      webOpenMode: string().required(m.report_form_required()),
      emailRecipients: array().of(
        string().email(m.report_form_invalid_email()),
      ),
      embedOwnerEmail: string().email(m.report_form_invalid_email()),
      enableSlackNotification: boolean(), // Needed to get the type for validation
      slackChannels: array().of(string()),
      slackUsers: array().of(string().email(m.report_form_invalid_email())),
      exportLimit: getExportLimitSchema(),
      columns: isCanvasReport
        ? array().of(string())
        : array().of(string()).min(1),
    })
      .test(
        "at-least-one-recipient",
        m.report_email_validation(),
        function (value) {
          // Check if at least one array has non-empty values
          const hasEmailRecipients =
            !!value.embedOwnerEmail?.trim() ||
            (value.emailRecipients
              ? value.emailRecipients.filter(Boolean).length > 0
              : false);
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
          // The members are unknown to users who cannot list them, so their recipients cannot be checked.
          const projectMembers = provider.projectMembersSet;
          if (!projectMembers) return true;

          return (
            values.emailRecipients?.every(
              (recipient) => !recipient || projectMembers.has(recipient),
            ) ?? true
          );
        },
      ),
  ) as ValidationAdapter<ReportValues>;

  const initialValues = $derived(
    props.mode === "create"
      ? getNewReportInitialFormValues(
          provider.userEmail,
          aggregationRequest,
          showEmbedOwnerEmail,
        )
      : props.mode === "create-canvas"
        ? getNewCanvasReportInitialFormValues(provider.userEmail)
        : getExistingReportInitialFormValues(
            props.reportSpec,
            provider.userEmail,
            aggregationRequest,
            showEmbedOwnerEmail,
          ),
  );

  // superForm registers lifecycle hooks, so it must be created once during init.
  // The initial values can change after init (e.g. the user's email loads), so reset the form to them.
  const { form, errors, enhance, submit, submitting, reset } = superForm(
    defaults(
      untrack(() => initialValues),
      schema,
    ),
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
  );
  $effect.pre(() => {
    const newState = initialValues;
    untrack(() => reset({ newState }));
  });

  let localError = $state<string | undefined>(undefined);
  const generalErrors = $derived(
    $errors._errors?.[0] ?? localError ?? ($mutation.error as any)?.message,
  );

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
      // Only sent if the field is shown, since the creator can't unsubscribe once their email is saved.
      // The admin service ignores it unless the creator edits the report, so other editors keep the saved email.
      embedOwnerEmail:
        isEmbedded && showEmbedOwnerEmail
          ? values.embedOwnerEmail.trim()
          : undefined,
    };

    // Until the specs load, the managers hold no filters, so the report would silently drop them.
    if (!expressionFilterManager.ready || !timeFilterManager.ready) {
      throw new Error(m.report_form_canvas_loading());
    }

    if (isCanvasReport && canvasBaseParams) {
      // The canvas state (filters, time range, tabs) the report will render with,
      // plus the PDF rendering options, which the export page reads back from this state.
      const stateParams = applyReportFiltersToParams(
        canvasBaseParams,
        expressionFilterManager,
        timeFilterManager,
      );
      stateParams.set("pdf_include_filters", String(values.pdfIncludeFilters));
      stateParams.set("pdf_all_tabs", String(values.pdfAllTabs));

      return {
        ...commonOptions,
        canvas: canvasName,
        exportFormat: V1ExportFormat.EXPORT_FORMAT_PDF,
        webOpenState: stateParams.toString(),
        // The selected filters are also baked into the report's security rules,
        // so magic-token recipients cannot query data beyond them.
        metricsViewFilters: getCanvasFilters(expressionFilterManager),
      };
    }

    const updatedAggregationRequest = buildAggregationRequest(
      aggregationRequest,
      [
        aggregationRequestWithTimeRange(timeFilterManager),
        aggregationRequestWithRowsAndColumns({
          exploreSpec,
          rows: values.rows,
          columns: values.columns,
          showTimeComparison: timeFilterManager.showComparison,
          selectedTimezone: timeFilterManager.timeZone,
          ephemeralMeasures,
        }),
      ],
    );
    updatedAggregationRequest.where =
      expressionFilterManager.topLevelJoiner.expr[
        dashboardConfigProvider.metricsViewsProvider.metricsViewNames[0]
      ];
    return {
      ...commonOptions,
      explore: exploreName,
      queryName: queryName,
      queryArgsJson: JSON.stringify(updatedAggregationRequest),
      exportLimit: values.exportLimit || undefined,
      exportIncludeHeader: values.exportIncludeHeader || false,
      exportAttach: values.exportAttach,
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
        org: provider.organization,
        project: provider.project,
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

      // Show report link only during create.
      // There only show when navigation is enabled in embed.
      const showReportLink =
        props.mode === "create" &&
        (!EmbedStore.getInstance() ||
          EmbedStore.getInstance()?.navigationEnabled);
      eventBus.emit("notification", {
        message:
          props.mode === "edit"
            ? m.report_form_edited_notification()
            : m.report_form_created_notification(),
        link: showReportLink
          ? {
              href: `${provider.pageBasePath}/reports`,
              text: m.report_form_go_to_reports(),
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
  <Dialog.Content
    class="min-w-[900px] max-h-screen overflow-auto"
    escapeKeydownBehavior="ignore"
  >
    <Dialog.Title>{m.report_form_schedule()}</Dialog.Title>

    <BaseScheduledReportForm
      formId={FORM_ID}
      data={form}
      {errors}
      {submit}
      {enhance}
      {exploreName}
      {canvasName}
      {dashboardConfigProvider}
      {expressionFilterManager}
      {timeFilterManager}
      {ephemeralMeasures}
      {showEmbedOwnerEmail}
      {provider}
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
