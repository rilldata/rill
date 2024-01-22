<script lang="ts">
  import { TabGroup, TabList } from "@rgossiaux/svelte-headlessui";
  import Tab from "./Tab.svelte";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";

  const StateManagers = getStateManagers();

  const { metricsViewName } = StateManagers;

  let currentTabIndex = 0;

  const tabs = [
    {
      label: "Explore",
    },
    {
      label: "Pivot",
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
          {tab.label}
        </Tab>
      {/each}
    </TabList>
  </TabGroup>
</div>
