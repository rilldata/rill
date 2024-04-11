<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { flip } from "svelte/animate";
  import { slide } from "svelte/transition";
  import { LIST_SLIDE_DURATION as duration } from "../../layout/config";
  import NavigationEntry from "../../layout/navigation/NavigationEntry.svelte";
  import NavigationHeader from "../../layout/navigation/NavigationHeader.svelte";
  import { runtime } from "../../runtime-client/runtime-store";
  import AddAssetButton from "../entity-management/AddAssetButton.svelte";
  import { getName } from "../entity-management/name-utils";
  import ChartMenuItems from "./ChartMenuItems.svelte";
  import { createChart } from "./createChart";
  import { useChartFileNames } from "./selectors";

  let showCharts = true;

  $: chartFileNames = useChartFileNames($runtime.instanceId);

  async function handleAddChart() {
    const newChartName = getName("chart", $chartFileNames.data ?? []);
    await createChart($runtime.instanceId, newChartName);
    await goto(`/files/charts/${newChartName}`);
  }
</script>

<div class="h-fit flex flex-col">
  <NavigationHeader bind:show={showCharts}>Charts</NavigationHeader>

  {#if showCharts}
    <ol transition:slide={{ duration }}>
      {#if $chartFileNames?.data}
        {#each $chartFileNames.data as chartName (chartName)}
          {@const open = $page.url.pathname === `/chart/${chartName}`}
          <li animate:flip={{ duration }} aria-label={chartName}>
            <NavigationEntry
              name={chartName}
              href={`/chart/${chartName}`}
              {open}
            >
              <ChartMenuItems slot="menu-items" {chartName} {open} />
            </NavigationEntry>
          </li>
        {/each}
      {/if}
      <AddAssetButton
        id="add-chart"
        label="Add chart"
        bold={false}
        on:click={handleAddChart}
      />
    </ol>
  {/if}
</div>
