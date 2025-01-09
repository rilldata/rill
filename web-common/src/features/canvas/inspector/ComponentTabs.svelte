<script lang="ts">
  import Tab from "@rilldata/web-common/features/dashboards/tab-bar/Tab.svelte";

  export let currentTab = "options";

  const tabs = [
    {
      tab: "options",
      label: "Options",
    },
    {
      tab: "filters",
      label: "Filters",
    },
  ];
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
