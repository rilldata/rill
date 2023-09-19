<script lang="ts">
  import { page } from "$app/stores";
  import StateManagersProvider from "@rilldata/web-common/features/dashboards/state-managers/StateManagersProvider.svelte";
  import { QueryClientProvider } from "@tanstack/svelte-query";
  import { createQueryClient } from "@rilldata/web-local/lib/svelte-query/globalQueryClient";
  import ExamplePivot from "@rilldata/web-common/features/dashboards/pivot/ExamplePivot.svelte";
  import TabGroup from "@rilldata/web-common/components/tab/TabGroup.svelte";
  import Tab from "@rilldata/web-common/components/tab/Tab.svelte";
  import BasicPivot from "./examples/basic/BasicPivot.svelte";
  import TimeDimensionDetails from "./examples/tdd/TimeDimensionDetails.svelte";

  $: metricViewName = $page.params.name;
  const queryClient = createQueryClient();
  const tabs = [
    { name: "Basic", component: BasicPivot },
    { name: "TDD styled", component: TimeDimensionDetails },
    { name: "Basic Async", component: ExamplePivot },
    { name: "Complex Async", component: ExamplePivot },
  ];

  let activeTab = tabs.at(0);
</script>

<QueryClientProvider client={queryClient}>
  <StateManagersProvider metricsViewName={metricViewName}>
    <TabGroup
      on:select={(event) => {
        activeTab = event.detail;
      }}
    >
      {#each tabs as tab}
        <Tab selected={activeTab === tab} value={tab}>{tab.name}</Tab>
      {/each}
    </TabGroup>

    {#if activeTab?.component}
      <div class="py-8">
        <svelte:component this={activeTab.component} />
      </div>
    {/if}
  </StateManagersProvider>
</QueryClientProvider>
