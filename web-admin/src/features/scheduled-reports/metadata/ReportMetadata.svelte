<script lang="ts">
  import { goto } from "$app/navigation";
  import { isNotFoundError } from "@rilldata/web-common/lib/errors";
  import MetadataList from "@rilldata/web-admin/features/scheduled-reports/metadata/MetadataList.svelte";
  import { extractNotifier } from "@rilldata/web-admin/features/scheduled-reports/metadata/notifiers-utils";
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { hasValidMetricsViewTimeRange } from "@rilldata/web-common/features/dashboards/selectors.ts";
  import { getMappedExploreUrl } from "@rilldata/web-common/features/explore-mappers/get-mapped-explore-url.ts";
  import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors";
  import ScheduledReportDialog from "@rilldata/web-common/features/scheduled-reports/ScheduledReportDialog.svelte";
  import { stripInternalReportParams } from "@rilldata/web-common/features/scheduled-reports/utils";
  import {
    ResourceKind,
    useResource,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import {
    getRuntimeServiceListResourcesQueryKey,
    V1ExportFormat,
    type V1Resource,
  } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { createAdminServiceDeleteReport } from "../../../client";
  import ProjectAccessControls from "../../projects/ProjectAccessControls.svelte";
  import {
    useIsReportCreatedByCode,
    useReport,
    useReportDashboardName,
  } from "../selectors";
  import MetadataLabel from "./MetadataLabel.svelte";
  import MetadataValue from "./MetadataValue.svelte";
  import ReportOwnerBlock from "./ReportOwnerBlock.svelte";
  import RunNowButton from "./RunNowButton.svelte";
  import {
    exportFormatToPrettyString,
    formatNextRunOn,
    formatRefreshSchedule,
  } from "./utils";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  export let organization: string;
  export let project: string;
  export let report: string;

  const runtimeClient = useRuntimeClient();

  $: reportQuery = useReport(runtimeClient, report);
  $: isReportCreatedByCode = useIsReportCreatedByCode(runtimeClient, report);

  $: reportSpec = $reportQuery.data?.resource?.report?.spec;
  $: isCanvasReport = !!reportSpec?.annotations?.canvas;

  // Get dashboard
  $: dashboardName = useReportDashboardName(runtimeClient, report);
  $: validSpecResp = useExploreValidSpec(
    runtimeClient,
    isCanvasReport ? "" : ($dashboardName.data ?? ""),
  );
  $: exploreSpec = $validSpecResp.data?.explore;
  $: canvasQuery = useResource<V1Resource>(
    runtimeClient,
    isCanvasReport ? ($dashboardName.data ?? "") : "",
    ResourceKind.Canvas,
  );
  $: canvasValidSpec = $canvasQuery.data?.canvas?.state?.validSpec;
  $: dashboardTitle = isCanvasReport
    ? canvasValidSpec?.displayName || $dashboardName.data
    : exploreSpec?.displayName || $dashboardName.data;
  $: dashboardDoesNotExist = isCanvasReport
    ? $canvasQuery.isError && isNotFoundError($canvasQuery.error)
    : $validSpecResp.isError && isNotFoundError($validSpecResp.error);

  $: exploreIsValid = hasValidMetricsViewTimeRange(
    runtimeClient,
    isCanvasReport ? "" : ($dashboardName.data ?? ""),
  );
  $: dashboardIsValid = isCanvasReport ? !!canvasValidSpec : $exploreIsValid;
  $: dashboardIsPending = isCanvasReport
    ? $canvasQuery.isPending
    : $validSpecResp.isPending;

  // Get human-readable frequency
  $: humanReadableFrequency = reportSpec?.refreshSchedule?.cron
    ? formatRefreshSchedule(reportSpec.refreshSchedule.cron)
    : "";

  $: emailNotifier = extractNotifier(reportSpec?.notifiers, "email");
  $: slackNotifier = extractNotifier(reportSpec?.notifiers, "slack");

  $: queryName =
    (reportSpec?.resolverProperties?.query_name as string | undefined) ??
    reportSpec?.queryName ??
    "";
  $: queryArgsJson =
    (reportSpec?.resolverProperties?.query_args_json as string | undefined) ??
    reportSpec?.queryArgsJson ??
    "";

  $: exploreUrl = getMappedExploreUrl(
    {
      exploreName: isCanvasReport ? "" : ($dashboardName.data ?? ""),
      queryName,
      queryArgsJson,
    },
    {
      exploreProtoState: reportSpec?.annotations?.web_open_state,
      forceOpenPivot: true,
    },
    {
      client: runtimeClient,
      organization,
      project,
    },
  );

  // Canvas reports link to the canvas page with the captured state applied.
  $: canvasUrl = (() => {
    if (!$dashboardName.data) return "";
    const path = `/${organization}/${project}/canvas/${$dashboardName.data}`;
    const search = stripInternalReportParams(
      new URLSearchParams(reportSpec?.annotations?.web_open_state ?? ""),
    ).toString();
    return search ? `${path}?${search}` : path;
  })();
  $: dashboardUrl = isCanvasReport ? canvasUrl : $exploreUrl;

  // Actions
  const queryClient = useQueryClient();
  const deleteReport = createAdminServiceDeleteReport();

  let showEditReportDialog = false;
  function handleEditReport() {
    showEditReportDialog = true;
  }

  async function handleDeleteReport() {
    await $deleteReport.mutateAsync({
      org: organization,
      project,
      name: $reportQuery.data.resource.meta.name.name,
    });
    queryClient.invalidateQueries({
      queryKey: getRuntimeServiceListResourcesQueryKey(
        runtimeClient.instanceId,
      ),
    });
    goto(`/${organization}/${project}/-/reports`);
  }
</script>

{#if reportSpec}
  <div class="flex flex-col gap-y-9 w-full max-w-full 2xl:max-w-[1200px]">
    <div class="flex flex-col gap-y-2">
      <!-- Header row 1 -->
      <div class="uppercase text-xs text-fg-secondary font-semibold">
        <!-- Author -->
        <ProjectAccessControls {organization} {project}>
          <svelte:fragment slot="manage-project">
            <ReportOwnerBlock
              {organization}
              {project}
              ownerId={reportSpec.annotations["admin_owner_user_id"]}
            />
          </svelte:fragment>
        </ProjectAccessControls>
        <!-- Format -->
        <span>
          {exportFormatToPrettyString(reportSpec.exportFormat)}
        </span>
        <!-- Limit (not applicable to PDF exports) -->
        {#if reportSpec.exportFormat !== V1ExportFormat.EXPORT_FORMAT_PDF}
          <span>
            • {reportSpec.exportLimit === "0"
              ? m.report_no_row_limit()
              : m.report_row_limit({ count: reportSpec.exportLimit })}
          </span>
        {/if}
      </div>
      <div class="flex gap-x-2 items-center">
        <h1
          class="text-fg-primary text-lg font-bold"
          aria-label={m.report_metadata_name_aria()}
        >
          {reportSpec.displayName}
        </h1>
        <div class="grow"></div>
        <RunNowButton {organization} {project} {report} />
        {#if !$isReportCreatedByCode.data}
          <DropdownMenu.Root>
            <DropdownMenu.Trigger>
              <IconButton ariaLabel={m.report_context_menu_aria()}>
                <ThreeDot size="16px" />
              </IconButton>
            </DropdownMenu.Trigger>
            <DropdownMenu.Content align="start">
              <DropdownMenu.Item
                onclick={handleEditReport}
                disabled={!dashboardIsValid}
              >
                {m.report_edit()}
              </DropdownMenu.Item>
              <DropdownMenu.Item onclick={handleDeleteReport}>
                {m.report_delete()}
              </DropdownMenu.Item>
            </DropdownMenu.Content>
          </DropdownMenu.Root>
        {/if}
      </div>
    </div>

    <!-- Three columns of metadata -->
    <div class="flex flex-wrap gap-x-16 gap-y-6">
      <!-- Dashboard -->
      <div
        class="flex flex-col gap-y-3"
        aria-label={m.report_metadata_dashboard_name_aria()}
      >
        {#if dashboardTitle}
          <MetadataLabel>{m.report_dashboard()}</MetadataLabel>
          <MetadataValue>
            {#if dashboardDoesNotExist}
              <div class="flex items-center gap-x-1">
                {dashboardTitle}
                <Tooltip distance={8}>
                  <CancelCircle size="16px" className="text-red-500" />
                  <TooltipContent slot="tooltip-content">
                    {m.report_dashboard_not_exist()}
                  </TooltipContent>
                </Tooltip>
              </div>
            {:else}
              <a href={dashboardUrl}>
                {dashboardTitle}
              </a>
            {/if}
          </MetadataValue>
        {:else}
          <MetadataLabel>{m.report_name_label()}</MetadataLabel>
          <MetadataValue>
            {$reportQuery.data?.resource?.meta?.name?.name}
          </MetadataValue>
        {/if}
      </div>

      <!-- Frequency -->
      <div
        class="flex flex-col gap-y-3"
        aria-label={m.report_metadata_schedule_aria()}
      >
        <MetadataLabel>{m.report_repeats()}</MetadataLabel>
        <MetadataValue>
          {humanReadableFrequency}
        </MetadataValue>
      </div>

      <!-- Next run -->
      <div class="flex flex-col gap-y-3">
        <MetadataLabel>{m.report_next_run()}</MetadataLabel>
        <MetadataValue>
          {formatNextRunOn(
            $reportQuery.data.resource.report.state.nextRunOn,
            reportSpec?.refreshSchedule?.timeZone,
          )}
        </MetadataValue>
      </div>
    </div>
    <!-- Slack recipients -->
    {#if slackNotifier}
      <MetadataList
        data={[...slackNotifier.channels, ...slackNotifier.users]}
        label={m.report_slack_recipients()}
      />
    {/if}

    <!-- Email recipients -->
    {#if emailNotifier}
      <MetadataList
        data={emailNotifier.recipients}
        label={m.report_email_recipients()}
      />
    {/if}
  </div>
{/if}

{#if reportSpec && dashboardIsValid && !dashboardIsPending && showEditReportDialog}
  <ScheduledReportDialog
    bind:open={showEditReportDialog}
    props={{
      mode: "edit",
      reportSpec,
    }}
  />
{/if}
