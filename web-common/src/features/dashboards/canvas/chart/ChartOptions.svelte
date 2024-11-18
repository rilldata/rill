<script lang="ts">
  import Tab from "@rilldata/web-common/features/dashboards/tab-bar/Tab.svelte";
  import AddMeasureDimensionButton from "../AddMeasureDimensionButton.svelte";
  import { chartConfig, updateAxis } from "./configStore";

  export let chartType = "bar";

  const tabs = [
    {
      label: "Data",
    },
    {
      label: "Display",
    },
    {
      label: "Axes",
    },
  ];

  let currentTabIndex = 0;

  function handleTabChange(index: number) {
    if (currentTabIndex === index) return;
    currentTabIndex = index;
  }
</script>

<div class="tabs">
  {#each tabs as { label }, i (label)}
    {@const selected = currentTabIndex === i}
    <Tab {selected} on:click={() => handleTabChange(i)}>
      <div class="flex gap-x-1 items-center group">
        {label}
      </div>
    </Tab>
  {/each}
</div>

{#if currentTabIndex === 0}
  <div class="chart-options">
    <div class="channel-name">X Axis:</div>
    {#if $chartConfig.data?.x?.field}
      {$chartConfig.data.x.field}
    {:else}
      Select a field
    {/if}
    <AddMeasureDimensionButton on:addField={(e) => updateAxis("x", e.detail)} />

    <div class="channel-name">Y Axis:</div>
    {#if $chartConfig.data?.y?.field}
      {$chartConfig.data.y.field}
    {:else}
      Select a field
    {/if}
    <AddMeasureDimensionButton on:addField={(e) => updateAxis("y", e.detail)} />

    <div class="channel-name">Color:</div>
    {#if $chartConfig.data?.color?.field}
      {$chartConfig.data.color.field}
    {:else}
      Select a field
    {/if}
    <AddMeasureDimensionButton
      on:addField={(e) => updateAxis("color", e.detail)}
    />
  </div>
{:else if currentTabIndex === 1}
  <div class="chart-options">
    <div class="channel-name">Display options placeholder</div>
  </div>
{:else if currentTabIndex === 2}
  <div class="chart-options">
    <div class="channel-name">Axes options placeholder</div>
  </div>
{/if}

<style lang="postcss">
  .chart-options {
    @apply px-4 py-2;
  }
  .tabs {
    @apply flex justify-center gap-x-4 h-8;
    @apply border-b border-slate-200;
  }
  .channel-name {
    @apply font-medium;
  }
</style>
