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
  import { get } from "svelte/store";
  import { runtime } from "../../runtime-client/runtime-store";
  import type TScheduledReportDialog from "../scheduled-reports/ScheduledReportDialog.svelte";

  export let disabled: boolean = false;
  export let workspace = false;
  export let label: string;
  export let includeScheduledReport = false;
  export let getQuery: (isScheduled: boolean) => V1Query | undefined;
  export let exploreName: string | undefined = undefined;

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
  const { reports } = featureFlags;

  async function handleExport(format: V1ExportFormat, includeHeader = false) {
    const result = await $exportDash.mutateAsync({
      instanceId: get(runtime).instanceId,
      data: {
        query: exportQuery,
        format,
        includeHeader,
      },
    });
    const downloadUrl = `${get(runtime).host}${result.downloadUrlPath}`;
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
      on:click={() => handleExport(V1ExportFormat.EXPORT_FORMAT_CSV)}
      disabled={!exportQuery}
    >
      Export as CSV
    </DropdownMenu.Item>
    <DropdownMenu.Item
      on:click={() => handleExport(V1ExportFormat.EXPORT_FORMAT_CSV, true)}
      disabled={!exportQuery}
    >
      Export as CSV with metadata
    </DropdownMenu.Item>
    <DropdownMenu.Item
      on:click={() => handleExport(V1ExportFormat.EXPORT_FORMAT_PARQUET)}
      disabled={!exportQuery}
    >
      Export as Parquet
    </DropdownMenu.Item>

    <DropdownMenu.Item
      on:click={() => handleExport(V1ExportFormat.EXPORT_FORMAT_XLSX)}
      disabled={!exportQuery}
    >
      Export as XLSX
    </DropdownMenu.Item>
    <DropdownMenu.Item
      on:click={() => handleExport(V1ExportFormat.EXPORT_FORMAT_XLSX, true)}
      disabled={!exportQuery}
    >
      Export as XLSX with metadata
    </DropdownMenu.Item>

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

<style lang="postcss">
  button {
    @apply h-6 px-1.5 py-px flex items-center gap-[3px] rounded-sm text-slate-600 pointer-events-auto;
  }

  button:hover {
    @apply bg-gray-200;
  }
</style>
