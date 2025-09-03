<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import Export from "@rilldata/web-common/components/icons/Export.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors.ts";
  import { getPivotExploreParams } from "@rilldata/web-common/features/exports/get-pivot-explore-params.ts";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import {
    createQueryServiceExport,
    V1ExportFormat,
    type V1Query,
  } from "@rilldata/web-common/runtime-client";
  import { get } from "svelte/store";
  import { runtime } from "../../runtime-client/runtime-store";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";

  export let disabled: boolean = false;
  export let workspace = false;
  export let label: string;
  export let includeScheduledReport = false;
  export let getQuery: (isScheduled: boolean) => V1Query | undefined;
  export let exploreName: string | undefined = undefined;

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

  $: ({ instanceId } = $runtime);
  $: exploreSpecQuery = useExploreValidSpec(instanceId, exploreName);
  $: exploreSpec = $exploreSpecQuery.data?.explore ?? {};

  async function handleExport(options: {
    format: V1ExportFormat;
    includeHeader?: boolean;
  }) {
    const { format, includeHeader = false } = options;
    const result = await $exportDash.mutateAsync({
      instanceId: get(runtime).instanceId,
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
    const downloadUrl = `${get(runtime).host}${result.downloadUrlPath}`;
    window.open(downloadUrl, "_self");
  }

  function createScheduledReport() {
    const pageState = get(page);
    const { organization, project } = pageState.params;
    const search = getPivotExploreParams(
      pageState.url.searchParams,
      exploreSpec,
    );
    void goto(
      `/${organization}/${project}/-/reports/create/${exploreName}?${search}`,
    );
  }
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
        on:click={createScheduledReport}
        disabled={!scheduledReportQuery}
      >
        Create scheduled report...
      </DropdownMenu.Item>
    {/if}
  </DropdownMenu.Content>
</DropdownMenu.Root>
