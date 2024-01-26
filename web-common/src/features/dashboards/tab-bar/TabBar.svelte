<script lang="ts">
  import { TabGroup, TabList } from "@rgossiaux/svelte-headlessui";
  import Chart from "@rilldata/web-common/components/icons/Chart.svelte";
  import Pivot from "@rilldata/web-common/components/icons/Pivot.svelte";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import Tab from "./Tab.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { featureFlags } from "../../feature-flags";

  const { pivot: pivotAllowed } = featureFlags;

  const StateManagers = getStateManagers();

  const {
    metricsViewName,
    selectors: {
      pivot: { showPivot },
    },
  } = StateManagers;

  $: currentTabIndex = $showPivot ? 1 : 0;

  const tabs = [
    {
      label: "Explore",
      icon: Chart,
    },
    {
      label: "Pivot",
      icon: Pivot,
    },
  ];

  function handleTabChange(event: CustomEvent) {
    const selectedTab = tabs[event.detail];

    if (selectedTab.label === "Pivot" && !$pivotAllowed) return;

    metricsExplorerStore.setPivotMode(
      $metricsViewName,
      selectedTab.label === "Pivot",
    );
  }
</script>

<div class="mr-4">
  <TabGroup defaultIndex={currentTabIndex} on:change={handleTabChange}>
    <TabList class="flex gap-x-4">
      {#each tabs as tab}
        {@const disabled = tab.label === "Pivot" && !$pivotAllowed}
        {#if disabled}
          <Tooltip>
            <TooltipContent slot="tooltip-content">Coming Soon</TooltipContent>
            <Tab {disabled}>
              <div class="flex gap-2 items-center">
                <svelte:component this={tab.icon} />
                {tab.label}
              </div>
            </Tab>
          </Tooltip>
        {:else}
          <Tab {disabled}>
            <div class="flex gap-2 items-center">
              <svelte:component this={tab.icon} />
              {tab.label}
            </div>
          </Tab>
        {/if}
      {/each}
    </TabList>
  </TabGroup>
</div>
