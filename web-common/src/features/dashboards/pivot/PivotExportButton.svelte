<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { builderActions, getAttrs } from "bits-ui";
  import { onMount } from "svelte";
  import CaretDownIcon from "../../../components/icons/CaretDownIcon.svelte";
  import {
    V1ExportFormat,
    createQueryServiceExport,
  } from "../../../runtime-client";
  import { useDashboard } from "../selectors";
  import { getStateManagers } from "../state-managers/state-managers";
  import exportPivot, { getPivotExportArgs } from "./pivot-export";

  export let includeScheduledReport: boolean;

  let active = false;
  let showScheduledReportDialog = false;

  const ctx = getStateManagers();
  const { runtime, metricsViewName } = ctx;
  const exportDash = createQueryServiceExport();

  $: metricsView = useDashboard($runtime.instanceId, $metricsViewName);

  async function handleExportPivot(format: V1ExportFormat) {
    await exportPivot({
      ctx,
      query: exportDash,
      format,
      timeDimension: $metricsView.data?.metricsView?.spec?.timeDimension,
    });
  }

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

  $: scheduledReportsQueryArgs = getPivotExportArgs(ctx);
</script>

<DropdownMenu.Root bind:open={active}>
  <DropdownMenu.Trigger asChild let:builder>
    <button
      class="h-6 px-1.5 py-px flex items-center gap-[3px] rounded-sm hover:bg-gray-200 text-gray-700"
      aria-label="Export pivot"
      {...getAttrs([builder])}
      on:click|preventDefault
      use:builderActions={{ builders: [builder] }}
    >
      Export
      <CaretDownIcon
        className="transition-transform {active && '-rotate-180'}"
        size="10px"
      />
    </button>
  </DropdownMenu.Trigger>
  <DropdownMenu.Content align="start">
    <DropdownMenu.Item
      on:click={async () => await handleExportPivot("EXPORT_FORMAT_CSV")}
    >
      Export as CSV
    </DropdownMenu.Item>
    <DropdownMenu.Item
      on:click={async () => await handleExportPivot("EXPORT_FORMAT_PARQUET")}
    >
      Export as Parquet
    </DropdownMenu.Item>
    <DropdownMenu.Item
      on:click={async () => await handleExportPivot("EXPORT_FORMAT_XLSX")}
    >
      Export as XLSX
    </DropdownMenu.Item>
    {#if includeScheduledReport}
      <DropdownMenu.Item on:click={() => (showScheduledReportDialog = true)}>
        Create scheduled report...
      </DropdownMenu.Item>
    {/if}
  </DropdownMenu.Content>
</DropdownMenu.Root>

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
