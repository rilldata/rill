<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import AddDataModal from "@rilldata/web-common/features/add-data/AddDataModal.svelte";
  import AddMetricsViewSubOption from "@rilldata/web-common/features/entity-management/add/AddMetricsViewSubOption.svelte";
  import CreateExploreDialog from "@rilldata/web-common/features/entity-management/add/CreateExploreDialog.svelte";
  import { createResourceAndNavigate } from "@rilldata/web-common/features/entity-management/add/new-files";
  import { resourceIconMapping } from "@rilldata/web-common/features/entity-management/resource-icon-mapping";
  import {
    ResourceKind,
    useFilteredResources,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { getScreenNameFromPage } from "@rilldata/web-common/features/file-explorer/telemetry";
  import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-common/metrics/service/MetricsTypes";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { PlusCircleIcon } from "lucide-svelte";

  let active = false;
  let showExploreDialog = false;
  let addDataModalOpen = false;
  let addDataConnector = "";
  let screenName = MetricsEventScreenName.Home;

  const runtimeClient = useRuntimeClient();

  $: metricsViewQuery = useFilteredResources(
    runtimeClient,
    ResourceKind.MetricsView,
  );
  $: metricsViews = $metricsViewQuery?.data ?? [];
</script>

<DropdownMenu.Root bind:open={active}>
  <DropdownMenu.Trigger>
    {#snippet child({ props })}
      <Button
        {...props}
        label="Add Asset"
        class="w-full"
        type="secondary"
        selected={active}
      >
        <PlusCircleIcon size="14px" />
        <div class="flex gap-x-1 items-center">
          Add
          <span class="transition-transform" class:-rotate-180={active}>
            <CaretDownIcon size="10px" />
          </span>
        </div>
      </Button>
    {/snippet}
  </DropdownMenu.Trigger>
  <DropdownMenu.Content align="start" class="w-[240px]">
    <AddMetricsViewSubOption
      onSelect={(connector) => {
        addDataModalOpen = true;
        addDataConnector = connector;
        screenName = getScreenNameFromPage();
      }}
    />
    <DropdownMenu.Separator />
    <DropdownMenu.Item
      aria-label="Add Explore Dashboard"
      class="flex gap-x-2"
      disabled={metricsViews.length === 0}
      onclick={() => {
        if (metricsViews.length === 1) {
          void createResourceAndNavigate(
            runtimeClient,
            ResourceKind.Explore,
            metricsViews.pop(),
          );
        } else {
          showExploreDialog = true;
        }
      }}
    >
      <div class="flex gap-x-2 items-center">
        <svelte:component
          this={resourceIconMapping[ResourceKind.Explore]}
          size="16px"
        />
        <div class="flex flex-col items-start">
          Explore dashboard
          {#if metricsViews.length === 0}
            <span class="text-fg-secondary text-xs">
              Requires a metrics view
            </span>
          {/if}
        </div>
      </div>
    </DropdownMenu.Item>
    <DropdownMenu.Item
      class="flex items-center gap-x-2"
      disabled={metricsViews.length === 0}
      onclick={() =>
        createResourceAndNavigate(runtimeClient, ResourceKind.Canvas)}
    >
      <div class="flex gap-x-2 items-center">
        <svelte:component
          this={resourceIconMapping[ResourceKind.Canvas]}
          size="16px"
        />
        <div class="flex flex-col items-start">
          Canvas dashboard
          {#if metricsViews.length === 0}
            <span class="text-fg-secondary text-xs">
              Requires a metrics view
            </span>
          {/if}
        </div>
      </div>
    </DropdownMenu.Item>
  </DropdownMenu.Content>
</DropdownMenu.Root>

<CreateExploreDialog bind:open={showExploreDialog} {metricsViews} />

<AddDataModal
  config={{
    medium: BehaviourEventMedium.Menu,
    space: MetricsEventSpace.LeftPanel,
    screen: screenName,
  }}
  bind:open={addDataModalOpen}
  connector={addDataConnector}
/>
