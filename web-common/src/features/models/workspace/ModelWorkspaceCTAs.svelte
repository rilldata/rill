<script lang="ts">
  import { goto } from "$app/navigation";
  import {
    Button,
    IconSpaceFixer,
  } from "@rilldata/web-common/components/button";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { WithTogglableFloatingElement } from "@rilldata/web-common/components/floating-element";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import Export from "@rilldata/web-common/components/icons/Export.svelte";
  import Forward from "@rilldata/web-common/components/icons/Forward.svelte";
  import { Menu, MenuItem } from "@rilldata/web-common/components/menu";
  import ResponsiveButtonText from "@rilldata/web-common/components/panel/ResponsiveButtonText.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import LocalAvatarButton from "@rilldata/web-common/features/authentication/LocalAvatarButton.svelte";
  import { removeLeadingSlash } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { createExportTableMutation } from "@rilldata/web-common/features/models/workspace/export-table";
  import {
    V1ExportFormat,
    V1ReconcileStatus,
    V1Resource,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { useGetDashboardsForModel } from "../../dashboards/selectors";
  import ModelRefreshButton from "../incremental/ModelRefreshButton.svelte";
  import CreateDashboardButton from "./CreateDashboardButton.svelte";

  export let resource: V1Resource | undefined;
  export let modelName: string;
  export let modelHasError = false;
  export let collapse = false;

  const exportModelMutation = createExportTableMutation();

  $: isModelIdle =
    resource?.meta?.reconcileStatus === V1ReconcileStatus.RECONCILE_STATUS_IDLE;

  $: dashboardsQuery = useGetDashboardsForModel($runtime.instanceId, modelName);

  $: availableDashboards = $dashboardsQuery.data ?? [];

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

<ModelRefreshButton {resource} {collapse} />

<DropdownMenu.Root>
  <DropdownMenu.Trigger asChild let:builder>
    <Button
      disabled={modelHasError || !isModelIdle}
      type="secondary"
      builders={[builder]}
    >
      <IconSpaceFixer pullLeft pullRight={collapse}>
        <Export />
      </IconSpaceFixer>

      <ResponsiveButtonText {collapse}>Export</ResponsiveButtonText>
      <CaretDownIcon />
    </Button>
  </DropdownMenu.Trigger>
  <DropdownMenu.Content align="end">
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

{#if availableDashboards?.length === 0}
  <CreateDashboardButton {collapse} hasError={modelHasError} {modelName} />
{:else if availableDashboards?.length === 1}
  <Tooltip distance={8} alignment="end">
    <Button
      type="primary"
      on:click={async () => {
        if (availableDashboards[0]?.meta?.filePaths?.[0]) {
          await goto(
            `/files/${removeLeadingSlash(availableDashboards[0].meta.filePaths[0])}`,
          );
        }
      }}
    >
      <IconSpaceFixer pullLeft pullRight={collapse}>
        <Forward />
      </IconSpaceFixer>
      <ResponsiveButtonText {collapse}>Go to dashboard</ResponsiveButtonText>
    </Button>
    <TooltipContent slot="tooltip-content">
      Go to the dashboard associated with this model
    </TooltipContent>
  </Tooltip>
{:else}
  <Tooltip distance={8} alignment="end">
    <WithTogglableFloatingElement
      let:toggleFloatingElement
      distance={8}
      alignment="end"
    >
      <Button on:click={toggleFloatingElement} type="primary">
        <IconSpaceFixer pullLeft pullRight={collapse}>
          <Forward />
        </IconSpaceFixer>
        <ResponsiveButtonText {collapse}>Go to dashboard</ResponsiveButtonText>
      </Button>
      <Menu
        dark
        slot="floating-element"
        let:toggleFloatingElement
        on:escape={toggleFloatingElement}
        on:click-outside={toggleFloatingElement}
      >
        {#each availableDashboards as resource (resource?.meta?.name?.name)}
          <MenuItem
            on:select={async () => {
              if (resource?.meta?.filePaths?.[0]) {
                await goto(
                  `/files/${removeLeadingSlash(resource.meta.filePaths[0])}`,
                );
                toggleFloatingElement();
              }
            }}
          >
            {resource?.meta?.name?.name ?? "Loading..."}
          </MenuItem>
        {/each}
      </Menu>
    </WithTogglableFloatingElement>
    <TooltipContent slot="tooltip-content">
      Go to one of {availableDashboards.length} dashboards associated with this model
    </TooltipContent>
  </Tooltip>
{/if}

<LocalAvatarButton />
