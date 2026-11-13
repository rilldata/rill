<script lang="ts">
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2";
  import type { ExpressionFilterManager } from "@rilldata/web-common/features/dashboards/filters/ExpressionFilterManager.svelte.ts";
  import { CanvasDashboardConfigProvider } from "@rilldata/web-common/features/dashboards/providers/DashboardConfigProvider.svelte.ts";
  import type { TimeFilterManager } from "@rilldata/web-common/features/dashboards/time-controls/TimeFilterManager.svelte.ts";
  import FiltersForm from "@rilldata/web-common/features/scheduled-reports/FiltersForm.svelte";
  import { createReportFilterManagers } from "@rilldata/web-common/features/scheduled-reports/utils.ts";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";

  /**
   * Test component that renders the canvas report filters the way `ScheduledReportDialog` and
   * `BaseScheduledReportForm` do: in-memory managers seeded with the report's canvas state.
   * `onManagersCreated` hands the managers to the test, which is where the dialog reads the saved filters from.
   */
  let {
    canvasName,
    canvasState,
    onManagersCreated,
  }: {
    canvasName: string;
    canvasState: string;
    onManagersCreated: (
      expressionFilterManager: ExpressionFilterManager,
      timeFilterManager: TimeFilterManager,
    ) => void;
  } = $props();

  const runtimeClient = useRuntimeClient();
  // svelte-ignore state_referenced_locally
  const dashboardConfigProvider = new CanvasDashboardConfigProvider(
    runtimeClient,
    canvasName,
  );
  // svelte-ignore state_referenced_locally
  const { expressionFilterManager, timeFilterManager } =
    createReportFilterManagers(
      runtimeClient,
      dashboardConfigProvider,
      false,
      new URLSearchParams(canvasState),
    );
  // svelte-ignore state_referenced_locally
  onManagersCreated(expressionFilterManager, timeFilterManager);
</script>

<!-- The app layout normally supplies the tooltip provider the time controls need. -->
<Tooltip.Provider>
  <FiltersForm
    {expressionFilterManager}
    {timeFilterManager}
    {dashboardConfigProvider}
    timeFiltersConfig={{ showGrainPicker: true }}
  />
</Tooltip.Provider>
