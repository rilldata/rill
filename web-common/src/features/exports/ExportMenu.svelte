<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import Export from "@rilldata/web-common/components/icons/Export.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import {
    createQueryServiceExport,
    V1ExportFormat,
    type V1Query,
  } from "@rilldata/web-common/runtime-client";
  import { builderActions, getAttrs } from "bits-ui";
  import { onMount } from "svelte";
  import { get } from "svelte/store";
  import { runtime } from "../../runtime-client/runtime-store";
  import type ScheduledReportDialog from "../scheduled-reports/ScheduledReportDialog.svelte";

  export let disabled: boolean = false;
  export let workspace = false;
  export let label: string;
  export let includeScheduledReport = false;
  export let query: V1Query;
  export let exploreName: string | undefined = undefined;
  export let metricsViewProto: string | undefined = undefined;

  let showScheduledReportDialog = false;
  let open = false;

  const exportDash = createQueryServiceExport();

  async function handleExport(format: V1ExportFormat) {
    const result = await $exportDash.mutateAsync({
      instanceId: get(runtime).instanceId,
      data: {
        query,
        format,
      },
    });
    const downloadUrl = `${get(runtime).host}${result.downloadUrlPath}`;
    window.open(downloadUrl, "_self");
  }

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
        <Button {label} {disabled} type="secondary" builders={[builder]} square>
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
      on:click={() => handleExport(V1ExportFormat.EXPORT_FORMAT_CSV)}
    >
      Export as CSV
    </DropdownMenu.Item>
    <DropdownMenu.Item
      on:click={() => handleExport(V1ExportFormat.EXPORT_FORMAT_PARQUET)}
    >
      Export as Parquet
    </DropdownMenu.Item>

    <DropdownMenu.Item
      on:click={() => handleExport(V1ExportFormat.EXPORT_FORMAT_XLSX)}
    >
      Export as XLSX
    </DropdownMenu.Item>

    {#if includeScheduledReport && query}
      <DropdownMenu.Item on:click={() => (showScheduledReportDialog = true)}>
        Create scheduled report...
      </DropdownMenu.Item>
    {/if}
  </DropdownMenu.Content>
</DropdownMenu.Root>

{#if includeScheduledReport && CreateScheduledReportDialog && showScheduledReportDialog && query}
  <svelte:component
    this={CreateScheduledReportDialog}
    {query}
    {metricsViewProto}
    {exploreName}
    bind:open={showScheduledReportDialog}
  />
{/if}

<style lang="postcss">
  button {
    @apply h-6 px-1.5 py-px flex items-center gap-[3px] rounded-sm text-gray-700 pointer-events-auto;
  }

  button:hover {
    @apply bg-gray-200;
  }
</style>
