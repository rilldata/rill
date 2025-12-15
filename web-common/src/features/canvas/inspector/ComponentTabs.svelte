<script lang="ts">
  import type { CanvasComponentType } from "@rilldata/web-common/features/canvas/components/types";
  import { isChartComponentType } from "@rilldata/web-common/features/canvas/components/util";
  import Tab from "@rilldata/web-common/features/dashboards/tab-bar/Tab.svelte";
  import { onMount } from "svelte";

  export let currentTab = "options";
  export let componentType: CanvasComponentType;
  export let hasFilters: boolean;

  const tabs = [
    {
      tab: "options",
      label: "Options",
    },
  ];

  $: if (hasFilters) {
    tabs.push({
      tab: "filters",
      label: "Filters",
    });
  }

  $: if (
    isChartComponentType(componentType) &&
    componentType !== "scatter_plot"
  ) {
    tabs.push({
      tab: "config",
      label: "Config",
    });
  }

  onMount(() => {
    currentTab = "options";
  });
</script>

<div class="mr-4">
  <div class="flex gap-x-2">
    {#each tabs as { tab, label } (tab)}
      {@const selected = tab === currentTab}
      <Tab {selected} on:click={() => (currentTab = tab)}>
        {label}
      </Tab>
    {/each}
  </div>
</div>
