<script lang="ts">
  import { WithTogglableFloatingElement } from "@rilldata/web-common/components/floating-element";
  import { Menu, MenuItem } from "@rilldata/web-common/components/menu";
  import { getDimensionTableExportArgs } from "@rilldata/web-common/features/dashboards/dimension-table/dimension-table-export-utils";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import {
    V1ExportFormat,
    createQueryServiceExport,
  } from "@rilldata/web-common/runtime-client";
  import { onMount } from "svelte";
  import CaretDownIcon from "../../../components/icons/CaretDownIcon.svelte";
  import exportToplist from "./export-toplist";

  export let includeScheduledReport: boolean;

  let exportMenuOpen = false;
  let showScheduledReportDialog = false;

  const ctx = getStateManagers();

  const exportDash = createQueryServiceExport();
  const handleExportTopList = async (format: V1ExportFormat) => {
    exportToplist({
      ctx,
      query: exportDash,
      format,
    });
  };

  // Only import the Scheduled Report dialog if in the Cloud context.
  // This ensures Rill Developer doesn't try and fail to import the admin-client.
  let CreateScheduledReportDialog;
  onMount(async () => {
    if (includeScheduledReport) {
      CreateScheduledReportDialog = (
        await import(
          "../../scheduled-reports/CreateScheduledReportDialog.svelte"
        )
      ).default;
    }
  });

  $: scheduledReportsQueryArgs = getDimensionTableExportArgs(ctx);
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
    let:toggleFloatingElement
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

<!-- Including `showScheduledReportDialog` in the conditional ensures we tear 
  down the form state when the dialog closes -->
{#if includeScheduledReport && CreateScheduledReportDialog && showScheduledReportDialog}
  <svelte:component
    this={CreateScheduledReportDialog}
    queryName="MetricsViewComparison"
    queryArgs={$scheduledReportsQueryArgs}
    open={showScheduledReportDialog}
    on:close={() => (showScheduledReportDialog = false)}
  />
{/if}
