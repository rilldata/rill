<script lang="ts">
  import { goto } from "$app/navigation";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import Add from "@rilldata/web-common/components/icons/Add.svelte";
  import MetricsViewIcon from "@rilldata/web-common/components/icons/MetricsViewIcon.svelte";
  import { removeLeadingSlash } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import { MetricsEventSpace } from "@rilldata/web-common/metrics/service/MetricsTypes";
  import {
    V1ReconcileStatus,
    type V1Resource,
  } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "../../../runtime-client/v2";
  import { useGetMetricsViewsForModel } from "../../dashboards/selectors";
  import ExportMenu from "../../exports/ExportMenu.svelte";
  import { useCreateMetricsViewFromTableUIAction } from "../../metrics-views/ai-generation/generateMetricsView";
  import NavigateOrDropdown from "../../metrics-views/NavigateOrDropdown.svelte";
  import ModelRefreshButton from "../incremental/ModelRefreshButton.svelte";
  import CreateDashboardButton from "./CreateDashboardButton.svelte";

  export let resource: V1Resource | undefined;
  export let modelName: string;
  export let hasResultTable = false;
  export let collapse = false;
  export let hasUnsavedChanges: boolean;
  export let connector: string;

  const runtimeClient = useRuntimeClient();

  $: ({ instanceId } = runtimeClient);
  $: isModelIdle =
    resource?.meta?.reconcileStatus === V1ReconcileStatus.RECONCILE_STATUS_IDLE;

  $: metricsViewsQuery = useGetMetricsViewsForModel(runtimeClient, modelName);

  $: availableMetricsViews = $metricsViewsQuery.data ?? [];

  $: createMetricsViewFromTable = useCreateMetricsViewFromTableUIAction(
    runtimeClient,
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
  disabled={!hasResultTable || !isModelIdle}
  workspace
  getQuery={() => {
    return {
      tableRowsRequest: {
        instanceId,
        connector,
        tableName: modelName,
      },
    };
  }}
/>

{#if availableMetricsViews?.length === 0}
  <CreateDashboardButton {collapse} {hasResultTable} {modelName} />
{:else}
  <DropdownMenu.Root>
    <DropdownMenu.Trigger asChild let:builder>
      <NavigateOrDropdown resources={availableMetricsViews} {builder} />
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
            <MetricsViewIcon size="16" />
            {resource?.meta?.name?.name ?? "Loading..."}
          </DropdownMenu.Item>
        {/each}
        <DropdownMenu.Separator />
        <DropdownMenu.Item
          on:click={async () => {
            if (!hasResultTable) return;
            await createMetricsViewFromTable();
          }}
          disabled={!hasResultTable}
        >
          <Add />
          Create metrics view
        </DropdownMenu.Item>
      </DropdownMenu.Content>
    {/if}
  </DropdownMenu.Root>
{/if}
