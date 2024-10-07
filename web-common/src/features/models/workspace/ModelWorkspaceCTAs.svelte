<script lang="ts">
  import { goto } from "$app/navigation";
  import {
    Button,
    IconSpaceFixer,
  } from "@rilldata/web-common/components/button";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import Export from "@rilldata/web-common/components/icons/Export.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { removeLeadingSlash } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { createExportTableMutation } from "@rilldata/web-common/features/models/workspace/export-table";
  import {
    V1ExportFormat,
    V1ReconcileStatus,
    type V1Resource,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { useGetMetricsViewsForModel } from "../../dashboards/selectors";
  import ModelRefreshButton from "../incremental/ModelRefreshButton.svelte";
  import CreateDashboardButton from "./CreateDashboardButton.svelte";

  export let resource: V1Resource | undefined;
  export let modelName: string;
  export let modelHasError = false;
  export let collapse = false;
  export let hasUnsavedChanges: boolean;

  const exportModelMutation = createExportTableMutation();

  let open = false;

  $: isModelIdle =
    resource?.meta?.reconcileStatus === V1ReconcileStatus.RECONCILE_STATUS_IDLE;

  $: metricsViewsQuery = useGetMetricsViewsForModel(
    $runtime.instanceId,
    modelName,
  );

  $: availableMetricsViews = $metricsViewsQuery.data ?? [];

  const onExport = async (format: V1ExportFormat) => {
    return $exportModelMutation.mutateAsync({
      data: {
        instanceId: $runtime.instanceId,
        format,
        tableName: modelName,
      },
    });
  };
</script>

<ModelRefreshButton {resource} {hasUnsavedChanges} />

<DropdownMenu.Root bind:open>
  <DropdownMenu.Trigger asChild let:builder>
    <Tooltip distance={8} suppress={open}>
      <Button
        disabled={modelHasError || !isModelIdle}
        type="secondary"
        builders={[builder]}
        square
      >
        <Export size="15px" />
      </Button>
      <TooltipContent slot="tooltip-content">Export model</TooltipContent>
    </Tooltip>
  </DropdownMenu.Trigger>
  <DropdownMenu.Content align="start">
    <DropdownMenu.Item
      on:click={() => onExport(V1ExportFormat.EXPORT_FORMAT_PARQUET)}
    >
      Export as Parquet
    </DropdownMenu.Item>
    <DropdownMenu.Item
      on:click={() => onExport(V1ExportFormat.EXPORT_FORMAT_CSV)}
    >
      Export as CSV
    </DropdownMenu.Item>
    <DropdownMenu.Item
      on:click={() => onExport(V1ExportFormat.EXPORT_FORMAT_XLSX)}
    >
      Export as XLSX
    </DropdownMenu.Item>
  </DropdownMenu.Content>
</DropdownMenu.Root>

{#if availableMetricsViews?.length === 0}
  <CreateDashboardButton {collapse} hasError={modelHasError} {modelName} />
{:else}
  <DropdownMenu.Root>
    <DropdownMenu.Trigger
      asChild
      let:builder
      on:click={async () => {
        if (availableMetricsViews[0]?.meta?.filePaths?.[0]) {
          await goto(
            `/files/${removeLeadingSlash(availableMetricsViews[0].meta.filePaths[0])}`,
          );
        }
      }}
    >
      <Tooltip distance={8} alignment="end">
        <Button builders={[builder]} type="secondary">
          Go to metrics view

          {#if availableMetricsViews.length > 1}
            <IconSpaceFixer pullRight>
              <CaretDownIcon />
            </IconSpaceFixer>
          {/if}
        </Button>

        <TooltipContent slot="tooltip-content">
          Go to one of {availableMetricsViews.length} metrics views associated with
          this model
        </TooltipContent>
      </Tooltip>
    </DropdownMenu.Trigger>

    {#if availableMetricsViews.length}
      <DropdownMenu.Content align="end">
        {#each availableMetricsViews as resource (resource?.meta?.name?.name)}
          <DropdownMenu.Item
            on:click={async () => {
              if (resource?.meta?.filePaths?.[0]) {
                await goto(
                  `/files/${removeLeadingSlash(resource.meta.filePaths[0])}`,
                );
              }
            }}
          >
            {resource?.meta?.name?.name ?? "Loading..."}
          </DropdownMenu.Item>
        {/each}
      </DropdownMenu.Content>
    {/if}
  </DropdownMenu.Root>
{/if}
