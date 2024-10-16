<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import Export from "@rilldata/web-common/components/icons/Export.svelte";
  import {
    V1ExportFormat,
    type V1MetricsViewAggregationRequest,
  } from "@rilldata/web-common/runtime-client";
  import { onMount } from "svelte";
  import type ScheduledReportDialog from "../scheduled-reports/ScheduledReportDialog.svelte";
  import { builderActions, getAttrs } from "bits-ui";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";

  export let disabled: boolean = false;
  export let workspace = false;
  export let label: string;
  export let includeScheduledReport = false;
  export let queryArgs: V1MetricsViewAggregationRequest | undefined = undefined;
  export let exploreName: string | undefined = undefined;
  export let metricsViewProto: string | undefined = undefined;
  export let onExport: (format: V1ExportFormat) => void;

  let showScheduledReportDialog = false;
  let open = false;

  // Only import the Scheduled Report dialog if in the Cloud context.
  // This ensures Rill Developer doesn't try and fail to import the admin-client.
  let CreateScheduledReportDialog: typeof ScheduledReportDialog;
  onMount(async () => {
    if (includeScheduledReport) {
      CreateScheduledReportDialog = (
        await import("../scheduled-reports/ScheduledReportDialog.svelte")
      ).default;
    }
  });
</script>

<DropdownMenu.Root bind:open>
  <DropdownMenu.Trigger asChild let:builder>
    {#if workspace}
      <Tooltip distance={8} suppress={open}>
        <Button {disabled} type="secondary" builders={[builder]} square>
          <Export size="15px" />
        </Button>
        <TooltipContent slot="tooltip-content">Export model</TooltipContent>
      </Tooltip>
    {:else}
      <button
        aria-label={label}
        use:builderActions={{ builders: [builder] }}
        {...getAttrs([builder])}
      >
        Export
        <CaretDownIcon
          className="transition-transform {open && '-rotate-180'}"
          size="10px"
        />
      </button>
    {/if}
  </DropdownMenu.Trigger>

  <DropdownMenu.Content align="start">
    <DropdownMenu.Item
      on:click={() => onExport(V1ExportFormat.EXPORT_FORMAT_CSV)}
    >
      Export as CSV
    </DropdownMenu.Item>
    <DropdownMenu.Item
      on:click={() => onExport(V1ExportFormat.EXPORT_FORMAT_PARQUET)}
    >
      Export as Parquet
    </DropdownMenu.Item>

    <DropdownMenu.Item
      on:click={() => onExport(V1ExportFormat.EXPORT_FORMAT_XLSX)}
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

{#if includeScheduledReport && CreateScheduledReportDialog && showScheduledReportDialog && queryArgs}
  <svelte:component
    this={CreateScheduledReportDialog}
    {queryArgs}
    {metricsViewProto}
    {exploreName}
    bind:open={showScheduledReportDialog}
  />
{/if}

<style lang="postcss">
  button {
    @apply h-6 px-1.5 py-px flex items-center gap-[3px] rounded-sm text-gray-700;
  }

  button:hover {
    @apply bg-gray-200;
  }
</style>
