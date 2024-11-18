<script lang="ts">
  import Tab from "@rilldata/web-common/features/dashboards/tab-bar/Tab.svelte";
  import AddMeasureDimensionButton from "../AddMeasureDimensionButton.svelte";

  export let chartType = "bar";

  let selectedXAxis = null;
  let selectedYAxis = null;
  function handleAddField(event) {
    const { detail: fieldName } = event;

    if (!selectedXAxis) {
      selectedXAxis = fieldName;
    } else {
      selectedYAxis = fieldName;
    }
  }

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
    {selectedXAxis || "Select a field"}
    <AddMeasureDimensionButton on:addField={handleAddField} />

    <div class="channel-name">Y Axis:</div>
    {selectedYAxis || "Select a field"}
    <AddMeasureDimensionButton on:addField={handleAddField} />
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
