<script lang="ts">
  import type { CanvasComponentType } from "@rilldata/web-common/features/canvas/components/types";
  import {
    getComponentRegistry,
    isChartComponentType,
  } from "@rilldata/web-common/features/canvas/components/util";
  import { hasComponentFilters } from "@rilldata/web-common/features/canvas/inspector/util";
  import Tab from "@rilldata/web-common/features/dashboards/tab-bar/Tab.svelte";
  import { onMount } from "svelte";

  export let currentTab = "options";
  export let componentType: CanvasComponentType;

  const componentsRegistry = getComponentRegistry();
  $: hasFilters = hasComponentFilters(componentsRegistry[componentType]);

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

  $: if (isChartComponentType(componentType)) {
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
