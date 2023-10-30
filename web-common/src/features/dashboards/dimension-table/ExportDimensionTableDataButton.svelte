<script lang="ts">
  import { WithTogglableFloatingElement } from "@rilldata/web-common/components/floating-element";
  import { Menu, MenuItem } from "@rilldata/web-common/components/menu";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import {
    createQueryServiceExport,
    V1ExportFormat,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { get } from "svelte/store";
  import CaretDownIcon from "../../../components/icons/CaretDownIcon.svelte";
  import { getQuerySortType } from "../leaderboard/leaderboard-utils";
  import { SortDirection } from "../proto-state/derived-types";
  import CreateScheduledReportModal from "../scheduled-reports/CreateScheduledReportModal.svelte";
  import { useDashboardStore } from "../stores/dashboard-stores";
  import exportToplist from "./export-toplist";

  export let metricViewName: string;

  let exportMenuOpen = false;
  let showScheduledReportDialog = false;

  const dashboardStore = useDashboardStore(metricViewName);
  const timeControlStore = useTimeControlStore(getStateManagers());

  const exportDash = createQueryServiceExport();
  const handleExportTopList = async (format: V1ExportFormat) => {
    exportToplist({
      metricViewName,
      query: exportDash,
      format,
      timeControlStore,
    });
  };
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
    on:click={(evt) => {
      evt.stopPropagation();
      toggleFloatingElement();
    }}
    class="h-6 px-1.5 py-px flex items-center gap-[3px] rounded-sm hover:bg-gray-200 text-gray-700"
  >
    Export
    <CaretDownIcon
      size="10px"
      className="transition-transform {exportMenuOpen && '-rotate-180'}"
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
    <!-- if Cloud -->
    <MenuItem
      on:select={() => {
        toggleFloatingElement();
        showScheduledReportDialog = true;
      }}
    >
      Create scheduled report...
    </MenuItem>
  </Menu>
</WithTogglableFloatingElement>

<CreateScheduledReportModal
  queryName="MetricsViewComparison"
  queryArgsJson={JSON.stringify({
    instanceId: get(runtime).instanceId,
    metricsViewName: metricViewName,
    dimension: {
      name: $dashboardStore.selectedDimensionName,
    },
    measures: $dashboardStore.selectedMeasureNames, // this is wrong; match dimension format
    timeRange: {
      start: $timeControlStore.timeStart,
      end: $timeControlStore.timeEnd,
    },
    comparisonTimeRange: {
      start: $timeControlStore.comparisonTimeStart,
      end: $timeControlStore.comparisonTimeEnd,
    },
    sort: [
      {
        name: $dashboardStore.leaderboardMeasureName,
        desc: $dashboardStore.sortDirection === SortDirection.DESCENDING,
        type: getQuerySortType($dashboardStore.dashboardSortType),
      },
    ],
    filter: $dashboardStore.filters,
    offset: "0",
  })}
  open={showScheduledReportDialog}
  on:close={() => (showScheduledReportDialog = false)}
/>
