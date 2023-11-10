<script lang="ts">
  import { WithTogglableFloatingElement } from "@rilldata/web-common/components/floating-element";
  import { Menu, MenuItem } from "@rilldata/web-common/components/menu";
  import { getDimensionTableExportArgs } from "@rilldata/web-common/features/dashboards/dimension-table/dimension-table-export-utils";
  import { useMetaQuery } from "@rilldata/web-common/features/dashboards/selectors/index";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import {
    createQueryServiceExport,
    V1ExportFormat,
  } from "@rilldata/web-common/runtime-client";
  import { onMount, SvelteComponent } from "svelte";
  import CaretDownIcon from "../../../components/icons/CaretDownIcon.svelte";
  import { useDashboardStore } from "../stores/dashboard-stores";
  import exportToplist from "./export-toplist";

  export let includeScheduledReport: boolean;
  export let metricViewName: string;

  let exportMenuOpen = false;
  let showScheduledReportDialog = false;

  const dashboardStore = useDashboardStore(metricViewName);
  const ctx = getStateManagers();
  const timeControlStore = useTimeControlStore(ctx);
  const metaQuery = useMetaQuery(ctx);

  const exportDash = createQueryServiceExport();
  const handleExportTopList = async (format: V1ExportFormat) => {
    exportToplist({
      metricViewName,
      query: exportDash,
      format,
      timeControlStore,
    });
  };

  // Only import the Scheduled Report modal if in the Cloud context
  // This ensures Rill Developer doesn't try and fail to import the admin-client
  let CreateScheduledReportModal: typeof SvelteComponent | undefined;
  onMount(async () => {
    if (includeScheduledReport) {
      CreateScheduledReportModal = (
        await import("../scheduled-reports/CreateScheduledReportModal.svelte")
      ).default;
    }
  });

  $: scheduledReportsQueryArgsJson = getDimensionTableExportArgs(
    metricViewName,
    $dashboardStore,
    $timeControlStore,
    $metaQuery.data
  );
</script>

<WithTogglableFloatingElement
  alignment="end"
  distance={8}
  let:toggleFloatingElement
  location="bottom"
  on:close={() => (exportMenuOpen = false)}
  on:open={() => (exportMenuOpen = true)}
>
  <button
    class="h-6 px-1.5 py-px flex items-center gap-[3px] rounded-sm hover:bg-gray-200 text-gray-700"
    on:click={(evt) => {
      evt.stopPropagation();
      toggleFloatingElement();
    }}
  >
    Export
    <CaretDownIcon
      className="transition-transform {exportMenuOpen && '-rotate-180'}"
      size="10px"
    />
  </button>
  <Menu
    minWidth=""
    on:click-outside={toggleFloatingElement}
    on:escape={toggleFloatingElement}
    slot="floating-element"
  >
    <MenuItem
      on:select={() => {
        toggleFloatingElement();
        handleExportTopList("EXPORT_FORMAT_CSV");
      }}
    >
      Export as CSV
    </MenuItem>
    <MenuItem
      on:select={() => {
        toggleFloatingElement();
        handleExportTopList("EXPORT_FORMAT_PARQUET");
      }}
    >
      Export as Parquet
    </MenuItem>
    <MenuItem
      on:select={() => {
        toggleFloatingElement();
        handleExportTopList("EXPORT_FORMAT_XLSX");
      }}
    >
      Export as XLSX
    </MenuItem>
    {#if includeScheduledReport}
      <MenuItem
        on:select={() => {
          toggleFloatingElement();
          showScheduledReportDialog = true;
        }}
      >
        Create scheduled report...
      </MenuItem>
    {/if}
  </Menu>
</WithTogglableFloatingElement>

{#if includeScheduledReport && CreateScheduledReportModal}
  <svelte:component
    this={CreateScheduledReportModal}
    queryName="MetricsViewComparison"
    queryArgsJson={scheduledReportsQueryArgsJson}
    open={showScheduledReportDialog}
    on:close={() => (showScheduledReportDialog = false)}
  />
{/if}
