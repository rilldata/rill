<script lang="ts">
  import { goto } from "$app/navigation";
  import { isNotFoundError } from "@rilldata/web-common/lib/errors";
  import EditAlert from "@rilldata/web-admin/features/alerts/EditAlert.svelte";
  import AlertFilterCriteria from "@rilldata/web-admin/features/alerts/metadata/AlertFilterCriteria.svelte";
  import AlertFilters from "@rilldata/web-admin/features/alerts/metadata/AlertFilters.svelte";
  import AlertOwnerBlock from "@rilldata/web-admin/features/alerts/metadata/AlertOwnerBlock.svelte";
  import { humaniseAlertSnoozeOption } from "@rilldata/web-admin/features/alerts/metadata/utils";
  import {
    useAlert,
    useAlertDashboardName,
    useIsAlertCreatedByCode,
  } from "@rilldata/web-admin/features/alerts/selectors";
  import ProjectAccessControls from "@rilldata/web-admin/features/projects/ProjectAccessControls.svelte";
  import DashboardMetadataValue from "@rilldata/web-admin/features/scheduled-reports/metadata/DashboardMetadataValue.svelte";
  import MetadataLabel from "@rilldata/web-admin/features/scheduled-reports/metadata/MetadataLabel.svelte";
  import MetadataList from "@rilldata/web-admin/features/scheduled-reports/metadata/MetadataList.svelte";
  import MetadataValue from "@rilldata/web-admin/features/scheduled-reports/metadata/MetadataValue.svelte";
  import { extractNotifier } from "@rilldata/web-admin/features/scheduled-reports/metadata/notifiers-utils";
  import {
    exportFormatToPrettyString,
    formatRefreshSchedule,
  } from "@rilldata/web-admin/features/scheduled-reports/metadata/utils.ts";
  import { IconButton } from "@rilldata/web-common/components/button";
  import {
    type AIAlertResolverProps,
    getAlertAggregationRequest,
    isAIAlertSpec,
  } from "@rilldata/web-common/features/alerts/ai-alert-utils";
  import { AlertNotifyRowsLimitDefault } from "@rilldata/web-common/features/alerts/form-utils";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import { hasValidMetricsViewTimeRange } from "@rilldata/web-common/features/dashboards/selectors.ts";
  import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors";
  import {
    getRuntimeServiceListResourcesQueryKey,
    V1ExportFormat,
  } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { createQuery, useQueryClient } from "@tanstack/svelte-query";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { EmbedStore } from "@rilldata/web-common/features/embeds/embed-store.ts";
  import { createMutationFromFactory } from "@rilldata/web-common/features/scheduled-reports/FormMetadataProvider.svelte.ts";
  import { alertDeleteMutationFactory } from "@rilldata/web-common/features/alerts/AlertFormMetadataProvider.ts";
  import { convertStateParamToUrlParams } from "@rilldata/web-common/features/dashboards/url-state/augment-url-params.ts";
  import { convertFilterParamToExpression } from "@rilldata/web-common/features/dashboards/url-state/filters/converters.ts";
  import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params.ts";
  import {
    getMappedExploreUrlQueryOptions,
    type MappedExploreUrlSource,
  } from "@rilldata/web-common/features/explore-mappers/url-mappers/get-mapped-url.ts";
  import { getDashboardPath } from "@rilldata/web-common/features/explore-mappers/utils";

  export let organization: string;
  export let project: string;
  export let alert: string;

  const runtimeClient = useRuntimeClient();
  // Embed users have no Rill session, so dashboard links must stay within the embed.
  const embedded = EmbedStore.isEmbedded();

  $: alertQuery = useAlert(runtimeClient, alert);
  $: isAlertCreatedByCode = useIsAlertCreatedByCode(runtimeClient, alert);

  // Get dashboard
  $: exploreName = useAlertDashboardName(runtimeClient, alert);
  $: validSpecResp = useExploreValidSpec(runtimeClient, $exploreName.data);
  $: exploreSpec = $validSpecResp.data?.explore;
  $: metricsViewName = exploreSpec?.metricsView;
  $: dashboardTitle = exploreSpec?.displayName || $exploreName.data;
  $: dashboardDoesNotExist =
    $validSpecResp.isError && isNotFoundError($validSpecResp.error);

  $: exploreIsValid = hasValidMetricsViewTimeRange(
    runtimeClient,
    $exploreName.data,
  );

  $: alertSpec = $alertQuery.data?.resource?.alert?.spec;

  // Get human-readable frequency
  $: humanReadableFrequency = alertSpec?.refreshSchedule?.cron
    ? formatRefreshSchedule(alertSpec.refreshSchedule.cron)
    : m.alert_whenever_data_refreshes();

  $: queryArgsJson =
    (alertSpec?.resolverProperties?.query_args_json as string) ||
    alertSpec?.queryArgsJson ||
    "{}";
  $: queryName =
    alertSpec?.queryName ||
    (alertSpec?.resolverProperties?.query_name as string);
  // An AI alert's resolver properties narrow its data the way a criteria alert's query does, and its condition replaces the criteria.
  $: isAIAlert = !!alertSpec && isAIAlertSpec(alertSpec);
  $: metricsViewAggregationRequest = alertSpec
    ? getAlertAggregationRequest(alertSpec, metricsViewName ?? "")
    : undefined;

  $: snoozeLabel = humaniseAlertSnoozeOption(alertSpec);

  $: emailNotifier = extractNotifier(alertSpec?.notifiers, "email");
  $: slackNotifier = extractNotifier(alertSpec?.notifiers, "slack");

  // The alert's request has no notion of in-list filters; the saved dashboard state does.
  $: dimensionsWithInlistFilter = getDimensionsWithInlistFilter(
    alertSpec?.annotations?.web_open_state,
  );
  function getDimensionsWithInlistFilter(webOpenState: string | undefined) {
    try {
      const filterParam = convertStateParamToUrlParams(
        "explore",
        webOpenState,
      ).get(ExploreStateURLParams.Filters);
      if (!filterParam) return [];
      return convertFilterParamToExpression(filterParam)
        .dimensionsWithInlistFilter;
    } catch {
      // An unparsable state only loses the in-list display of the filters.
      return [];
    }
  }

  // AI alerts link to the explore they analyze, with the alert's resolver properties applied.
  $: exploreUrlSource = (
    isAIAlert
      ? {
          type: "ai",
          exploreName: $exploreName.data ?? "",
          props: alertSpec?.resolverProperties as AIAlertResolverProps,
        }
      : {
          type: "query",
          request: {
            exploreName: $exploreName.data ?? "",
            queryName,
            queryArgsJson,
          },
          options: {
            exploreProtoState: alertSpec?.annotations?.web_open_state,
          },
        }
  ) as MappedExploreUrlSource;
  $: exploreUrlQueryOptions = getMappedExploreUrlQueryOptions(
    runtimeClient,
    exploreUrlSource,
    { organization, project, embedded },
  );
  // Until the explore name loads, the url would fall back to the metrics view's name.
  $: exploreUrlQuery = createQuery({
    ...exploreUrlQueryOptions,
    enabled: exploreUrlQueryOptions.enabled && $exploreName.data !== undefined,
  });
  // The link is shown once the url is mapped; if mapping fails it still opens the explore, without the alert's state.
  $: exploreUrl = $exploreUrlQuery.isError
    ? getDashboardPath(
        { organization, project, embedded },
        "explore",
        $exploreName.data ?? "",
      )
    : $exploreUrlQuery.data;

  // Actions
  const queryClient = useQueryClient();
  const deleteAlert = createMutationFromFactory(alertDeleteMutationFactory);

  async function handleDeleteAlert() {
    await $deleteAlert.mutateAsync({
      org: organization,
      project,
      name: $alertQuery.data.resource.meta.name.name,
    });
    await queryClient.invalidateQueries({
      queryKey: getRuntimeServiceListResourcesQueryKey(
        runtimeClient.instanceId,
      ),
    });
    // goto only after invalidate is complete
    goto(embedded ? "/-/embed/alerts" : `/${organization}/${project}/-/alerts`);
  }
</script>

{#if alertSpec}
  <div class="flex flex-col gap-y-9 w-full max-w-full 2xl:max-w-[1200px]">
    <div class="flex flex-col gap-y-2">
      <!-- Header row 1 -->
      <div class="uppercase text-xs text-fg-secondary font-semibold">
        <!-- Author -->
        <ProjectAccessControls {organization} {project}>
          <svelte:fragment slot="manage-project">
            {#if $alertQuery.data}
              <AlertOwnerBlock
                {organization}
                {project}
                ownerId={alertSpec.annotations["admin_owner_user_id"]}
                embedOwnerId={alertSpec.annotations?.["admin_embed_owner_id"]}
              />
            {/if}
          </svelte:fragment>
        </ProjectAccessControls>
      </div>
      <div class="flex gap-x-2 items-center">
        <h1
          class="text-fg-primary text-lg font-bold"
          aria-label={m.alert_metadata_name_aria()}
        >
          {alertSpec.displayName}
        </h1>
        <div class="grow"></div>
        {#if !$isAlertCreatedByCode.data}
          <EditAlert {alertSpec} disabled={!$exploreIsValid} />
          <DropdownMenu.Root>
            <DropdownMenu.Trigger>
              <IconButton ariaLabel={m.alert_context_menu_aria()}>
                <ThreeDot size="16px" />
              </IconButton>
            </DropdownMenu.Trigger>
            <DropdownMenu.Content align="start">
              <DropdownMenu.Item onclick={handleDeleteAlert}>
                {m.alert_delete()}
              </DropdownMenu.Item>
            </DropdownMenu.Content>
          </DropdownMenu.Root>
        {/if}
      </div>
    </div>

    <!-- Five columns of metadata -->
    <div class="flex flex-wrap gap-x-16 gap-y-6">
      <!-- Dashboard -->
      <div
        class="flex flex-col gap-y-3"
        aria-label={m.alert_metadata_dashboard_name_aria()}
      >
        {#if dashboardTitle}
          <MetadataLabel>{m.alert_dashboard()}</MetadataLabel>
          <DashboardMetadataValue
            title={dashboardTitle}
            href={exploreUrl}
            state={dashboardDoesNotExist
              ? "missing"
              : $exploreUrlQuery.isError
                ? "stateLoadFailed"
                : "ok"}
            notExistMessage={m.alert_dashboard_not_exist()}
            stateLoadFailedMessage={m.alert_dashboard_state_load_failed()}
            openWithoutStateLabel={m.alert_dashboard_open_without_filters()}
          />
        {:else}
          <MetadataLabel>{m.alert_name_label()}</MetadataLabel>
          <MetadataValue>
            {$alertQuery.data?.resource?.meta?.name?.name}
          </MetadataValue>
        {/if}
      </div>

      <!-- Split by dimension -->
      <div
        class="flex flex-col gap-y-3"
        aria-label={m.alert_metadata_split_by_dimension_aria()}
      >
        <MetadataLabel>{m.alert_split_by_dimension()}</MetadataLabel>
        <MetadataValue>
          {metricsViewAggregationRequest?.dimensions?.[0]?.name ??
            m.alert_none()}
        </MetadataValue>
      </div>

      <!-- Rows in notifications, which only matters when a criteria alert is split by a dimension -->
      {#if metricsViewAggregationRequest?.dimensions?.length && !isAIAlert}
        <div class="flex flex-col gap-y-3">
          <MetadataLabel>{m.alert_rows_in_notifications()}</MetadataLabel>
          <MetadataValue>
            {alertSpec.notifyRowsLimit || AlertNotifyRowsLimitDefault}
          </MetadataValue>
        </div>
      {/if}

      <!-- Schedule: TODO: change based on non UI settings -->
      <div
        class="flex flex-col gap-y-3"
        aria-label={m.alert_metadata_schedule_aria()}
      >
        <MetadataLabel>{m.alert_schedule()}</MetadataLabel>
        <MetadataValue>{humanReadableFrequency}</MetadataValue>
      </div>

      <!-- Snooze -->
      <div class="flex flex-col gap-y-3">
        <MetadataLabel>{m.alert_snooze()}</MetadataLabel>
        <MetadataValue>{snoozeLabel}</MetadataValue>
      </div>

      <!-- Email attachment -->
      {#if alertSpec.exportAttach}
        <div class="flex flex-col gap-y-3">
          <MetadataLabel>{m.alert_email_attachment()}</MetadataLabel>
          <MetadataValue>
            {exportFormatToPrettyString(
              alertSpec.exportFormat ?? V1ExportFormat.EXPORT_FORMAT_CSV,
            )} • {!alertSpec.exportLimit || alertSpec.exportLimit === "0"
              ? m.report_no_row_limit()
              : m.report_row_limit({ count: alertSpec.exportLimit })}
          </MetadataValue>
        </div>
      {/if}
    </div>

    <!-- Filters -->
    <AlertFilters
      {metricsViewName}
      filters={metricsViewAggregationRequest?.where}
      {dimensionsWithInlistFilter}
      timeRange={metricsViewAggregationRequest?.timeRange}
      comparisonTimeRange={metricsViewAggregationRequest?.comparisonTimeRange}
    />

    <!-- Criteria, or an AI alert's condition -->
    {#if isAIAlert}
      <div class="flex flex-col gap-y-3">
        <MetadataLabel>{m.alert_condition()}</MetadataLabel>
        <MetadataValue>
          {alertSpec.resolverProperties?.prompt}
        </MetadataValue>
      </div>
    {:else}
      <AlertFilterCriteria
        filters={metricsViewAggregationRequest?.having}
        comparisonTimeRange={metricsViewAggregationRequest?.comparisonTimeRange}
      />
    {/if}

    <!-- Slack notification -->
    {#if slackNotifier}
      <MetadataList
        data={[...slackNotifier.channels, ...slackNotifier.users]}
        label={m.alert_slack_notifications()}
      />
    {/if}

    <!-- Email notifications -->
    {#if emailNotifier}
      <MetadataList
        data={emailNotifier.recipients}
        label={m.alert_email_notifications()}
      />
    {/if}
  </div>
{/if}
