<script lang="ts">
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2";
  import {
    hasValidMetricsViewTimeRange,
    useMetricsViewTimeRange,
  } from "@rilldata/web-common/features/dashboards/selectors.ts";
  import ScheduledReportDialog from "@rilldata/web-common/features/scheduled-reports/ScheduledReportDialog.svelte";
  import { getDashboardNameFromReport } from "@rilldata/web-common/features/scheduled-reports/utils";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import {
    createRuntimeServiceListResources,
    type V1ReportSpec,
  } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";

  /**
   * Test component that opens the edit dialog of a report the way the report page does.
   * The page has loaded everything the dialog reads before the dialog can be opened,
   * so the dialog starts from a full cache; the queries below are the ones the page runs.
   *
   * Mirrors `ReportMetadata`; keep the two in step.
   */
  let {
    reportSpec,
    metricsViewName,
  }: { reportSpec: V1ReportSpec; metricsViewName: string } = $props();

  const runtimeClient = useRuntimeClient();

  // The report is fixed for the lifetime of a test, so reading the props once at init is enough.
  // The page only renders the dialog for a valid explore.
  // svelte-ignore state_referenced_locally
  const exploreIsValid = hasValidMetricsViewTimeRange(
    runtimeClient,
    getDashboardNameFromReport(reportSpec),
  );
  // The link to the dashboard needs the time range of the report's metrics view.
  // svelte-ignore state_referenced_locally
  const timeRange = useMetricsViewTimeRange(
    runtimeClient,
    metricsViewName,
    undefined,
    queryClient,
  );
  // The project nav lists every resource for its status indicator.
  const resources = createRuntimeServiceListResources(runtimeClient, {});
  let pageLoaded = $derived(
    $exploreIsValid && $timeRange.isSuccess && $resources.isSuccess,
  );

  let open = $state(true);
</script>

<!-- The app layout normally supplies the tooltip provider the time controls need. -->
<Tooltip.Provider>
  {#if pageLoaded && open}
    <ScheduledReportDialog bind:open props={{ mode: "edit", reportSpec }} />
  {/if}
</Tooltip.Provider>
