<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { flip } from "svelte/animate";
  import { slide } from "svelte/transition";
  import { LIST_SLIDE_DURATION } from "../../layout/config";
  import NavigationEntry from "../../layout/navigation/NavigationEntry.svelte";
  import NavigationHeader from "../../layout/navigation/NavigationHeader.svelte";
  import { runtime } from "../../runtime-client/runtime-store";
  import AddAssetButton from "../entity-management/AddAssetButton.svelte";
  import { getName } from "../entity-management/name-utils";
  import {
    ResourceKind,
    useFilteredResourceNames,
  } from "../entity-management/resource-selectors";
  import ChartMenuItems from "./ChartMenuItems.svelte";
  import { createChart } from "./createChart";

  $: chartNames = useFilteredResourceNames(
    $runtime.instanceId,
    ResourceKind.Chart,
  );

  let showCharts = true;

  async function handleAddChart() {
    const newChartName = getName("chart", $chartNames.data ?? []);
    await createChart($runtime.instanceId, newChartName);
    await goto(`/chart/${newChartName}`);
  }
</script>

<NavigationHeader bind:show={showCharts} toggleText="charts">
  Charts
</NavigationHeader>

{#if showCharts}
  <div
    class="pb-3 max-h-96 overflow-auto"
    transition:slide={{ duration: LIST_SLIDE_DURATION }}
  >
    {#if $chartNames?.data}
      {#each $chartNames.data as chartName (chartName)}
        <div
          animate:flip={{ duration: 200 }}
          out:slide|global={{ duration: LIST_SLIDE_DURATION }}
        >
          <NavigationEntry
            name={chartName}
            href={`/chart/${chartName}`}
            open={$page.url.pathname === `/chart/${chartName}`}
            expandable={false}
          >
            <svelte:fragment slot="menu-items">
              <ChartMenuItems {chartName} />
            </svelte:fragment>
          </NavigationEntry>
        </div>
      {/each}
    {/if}
    <AddAssetButton
      id="add-chart"
      label="Add chart"
      bold={false}
      on:click={handleAddChart}
    />
  </div>
{/if}
