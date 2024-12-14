<script lang="ts">
  import Tab from "@rilldata/web-common/features/dashboards/tab-bar/Tab.svelte";

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
    <!-- <FieldSelectorDropdown
      label="X axis"
      id="x-axis"
      type="dimension"
      selectedItem={$chartConfig.data?.x?.field}
      onSelect={(field) => updateAxis("x", field)}
    />

    <FieldSelectorDropdown
      label="Y axis"
      id="y-axis"
      type="measure"
      selectedItem={$chartConfig.data?.y?.field}
      onSelect={(field) => updateAxis("y", field)}
    />

    <FieldSelectorDropdown
      label="Color"
      id="color-axis"
      type="dimension"
      selectedItem={$chartConfig.data?.color?.field}
      onSelect={(field) => updateAxis("color", field)}
    /> -->
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
    @apply flex flex-col px-4 py-2 gap-y-4;
  }
  .tabs {
    @apply flex justify-center gap-x-4 h-8;
    @apply border-b border-slate-200;
  }
  .channel-name {
    @apply font-medium;
  }
</style>
