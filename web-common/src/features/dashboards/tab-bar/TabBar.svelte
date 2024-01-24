<script lang="ts">
  import { TabGroup, TabList } from "@rgossiaux/svelte-headlessui";
  import Tab from "./Tab.svelte";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import Chart from "@rilldata/web-common/components/icons/Chart.svelte";
  import Pivot from "@rilldata/web-common/components/icons/Pivot.svelte";

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
    console.log(`Switching to tab: ${selectedTab.label}`);

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
        <Tab>
          <div class="flex gap-2 items-center">
            <svelte:component this={tab.icon} />
            {tab.label}
          </div>
        </Tab>
      {/each}
    </TabList>
  </TabGroup>
</div>
