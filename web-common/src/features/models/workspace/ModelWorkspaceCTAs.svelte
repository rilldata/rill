<script lang="ts">
  import { goto } from "$app/navigation";
  import { Button } from "@rilldata/web-common/components/button";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import Add from "@rilldata/web-common/components/icons/Add.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import MetricsViewIcon from "@rilldata/web-common/components/icons/MetricsViewIcon.svelte";
  import { removeLeadingSlash } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import { MetricsEventSpace } from "@rilldata/web-common/metrics/service/MetricsTypes";
  import {
    V1ReconcileStatus,
    type V1Resource,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { useGetMetricsViewsForModel } from "../../dashboards/selectors";
  import { resourceColorMapping } from "../../entity-management/resource-icon-mapping";
  import { ResourceKind } from "../../entity-management/resource-selectors";
  import ExportMenu from "../../exports/ExportMenu.svelte";
  import { useCreateMetricsViewFromTableUIAction } from "../../metrics-views/ai-generation/generateMetricsView";
  import ModelRefreshButton from "../incremental/ModelRefreshButton.svelte";
  import CreateDashboardButton from "./CreateDashboardButton.svelte";

  export let resource: V1Resource | undefined;
  export let modelName: string;
  export let modelHasError = false;
  export let collapse = false;
  export let hasUnsavedChanges: boolean;
  export let connector: string;

  $: ({ instanceId } = $runtime);
  $: isModelIdle =
    resource?.meta?.reconcileStatus === V1ReconcileStatus.RECONCILE_STATUS_IDLE;

  $: metricsViewsQuery = useGetMetricsViewsForModel(instanceId, modelName);

  $: availableMetricsViews = $metricsViewsQuery.data ?? [];

  $: createMetricsViewFromTable = useCreateMetricsViewFromTableUIAction(
    instanceId,
    connector,
    "",
    "",
    modelName,
    false,
    BehaviourEventMedium.Menu,
    MetricsEventSpace.LeftPanel,
  );
</script>

<ModelRefreshButton {resource} {hasUnsavedChanges} />

<ExportMenu
  label="Export model data"
  disabled={modelHasError || !isModelIdle}
  workspace
  query={{
    tableRowsRequest: {
      instanceId,
      tableName: modelName,
    },
  }}
/>

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
      <Button builders={[builder]} type="secondary">
        Go to metrics view
        <CaretDownIcon />
      </Button>
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
            <MetricsViewIcon
              size="16"
              color={resourceColorMapping[ResourceKind.MetricsView]}
            />
            {resource?.meta?.name?.name ?? "Loading..."}
          </DropdownMenu.Item>
        {/each}
        <DropdownMenu.Separator />
        <DropdownMenu.Item
          on:click={async () => {
            await createMetricsViewFromTable();
          }}
        >
          <Add />
          Create metrics view
        </DropdownMenu.Item>
      </DropdownMenu.Content>
    {/if}
  </DropdownMenu.Root>
{/if}
