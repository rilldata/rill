<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import Export from "@rilldata/web-common/components/icons/Export.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import {
    createQueryServiceExport,
    V1ExportFormat,
    type V1Query,
  } from "@rilldata/web-common/runtime-client";
  import { onMount } from "svelte";

  import type TScheduledReportDialog from "../scheduled-reports/ScheduledReportDialog.svelte";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import httpClient from "@rilldata/web-common/runtime-client/http-client";

  export let disabled: boolean = false;
  export let workspace = false;
  export let label: string;
  export let includeScheduledReport = false;
  export let getQuery: (isScheduled: boolean) => V1Query | undefined;
  export let exploreName: string | undefined = undefined;

  const instanceId = httpClient.getInstanceId();

  let showScheduledReportDialog = false;
  let open = false;
  let exportQuery: V1Query | undefined;
  let scheduledReportQuery: V1Query | undefined;

  // Get the query when the dialog is opened.
  // (Note: it might be better to pass pre-computed queries into the `ExportMenu` component.)
  $: if (open) {
    exportQuery = getQuery(false);
    scheduledReportQuery = getQuery(true);
  }

  const exportDash = createQueryServiceExport();
  const { reports, adminServer, exportHeader } = featureFlags;

  async function handleExport(options: {
    format: V1ExportFormat;
    includeHeader?: boolean;
  }) {
    const { format, includeHeader = false } = options;
    const result = await $exportDash.mutateAsync({
      instanceId,
      data: {
        query: exportQuery,
        format,
        includeHeader,
        // Include metadata for CSV/XLSX exports in Cloud context.
        ...(includeHeader &&
          $adminServer && {
            originDashboard: { name: exploreName, kind: ResourceKind.Explore },
            origin_url: window.location.href,
          }),
      },
    });
    const downloadUrl = `${httpClient.getHost()}${result.downloadUrlPath}`;
    window.open(downloadUrl, "_self");
  }

  // Only import the Scheduled Report dialog if in the Cloud context.
  // This ensures Rill Developer doesn't try and fail to import the admin-client.
  let ScheduledReportDialog: typeof TScheduledReportDialog;
  onMount(async () => {
    if (includeScheduledReport) {
      ({ default: ScheduledReportDialog } = await import(
        "../scheduled-reports/ScheduledReportDialog.svelte"
      ));
    }
  });
</script>

<DropdownMenu.Root bind:open>
  <DropdownMenu.Trigger asChild let:builder>
    {#if workspace}
      <Tooltip distance={8} suppress={open}>
        <Button {label} {disabled} type="secondary" builders={[builder]} square>
          <Export size="15px" />
        </Button>
        <TooltipContent slot="tooltip-content">Export model</TooltipContent>
      </Tooltip>
    {:else}
      <Button {label} {disabled} type="toolbar" builders={[builder]}>
        <Export size="15px" />
        Export
        <CaretDownIcon
          className="transition-transform {open && '-rotate-180'}"
          size="10px"
        />
      </Button>
    {/if}
  </DropdownMenu.Trigger>

  <DropdownMenu.Content align="start">
    <DropdownMenu.Item
      on:click={() =>
        handleExport({ format: V1ExportFormat.EXPORT_FORMAT_CSV })}
      disabled={!exportQuery}
    >
      Export as CSV
    </DropdownMenu.Item>
    {#if !workspace && $exportHeader}
      <DropdownMenu.Item
        on:click={() =>
          handleExport({
            format: V1ExportFormat.EXPORT_FORMAT_CSV,
            includeHeader: true,
          })}
        disabled={!exportQuery}
      >
        Export as CSV with metadata
      </DropdownMenu.Item>
    {/if}
    <DropdownMenu.Item
      on:click={() =>
        handleExport({ format: V1ExportFormat.EXPORT_FORMAT_PARQUET })}
      disabled={!exportQuery}
    >
      Export as Parquet
    </DropdownMenu.Item>

    <DropdownMenu.Item
      on:click={() =>
        handleExport({ format: V1ExportFormat.EXPORT_FORMAT_XLSX })}
      disabled={!exportQuery}
    >
      Export as XLSX
    </DropdownMenu.Item>
    {#if !workspace && $exportHeader}
      <DropdownMenu.Item
        on:click={() =>
          handleExport({
            format: V1ExportFormat.EXPORT_FORMAT_XLSX,
            includeHeader: true,
          })}
        disabled={!exportQuery}
      >
        Export as XLSX with metadata
      </DropdownMenu.Item>
    {/if}
    {#if includeScheduledReport && $reports && exploreName}
      <DropdownMenu.Item
        on:click={() => (showScheduledReportDialog = true)}
        disabled={!scheduledReportQuery}
      >
        Create scheduled report...
      </DropdownMenu.Item>
    {/if}
  </DropdownMenu.Content>
</DropdownMenu.Root>

{#if includeScheduledReport && ScheduledReportDialog && showScheduledReportDialog && scheduledReportQuery && exploreName}
  <svelte:component
    this={ScheduledReportDialog}
    bind:open={showScheduledReportDialog}
    props={{
      mode: "create",
      query: scheduledReportQuery,
      exploreName,
    }}
  />
{/if}
