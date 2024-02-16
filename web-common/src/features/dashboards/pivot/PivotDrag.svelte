<script context="module" lang="ts">
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import DragList from "./DragList.svelte";
  import type { PivotSidebarSection, PivotChipData } from "./types";
</script>

<script lang="ts">
  export let title: PivotSidebarSection;
  export let items: PivotChipData[];
  export let collapsed = false;

  function toggleCollapse() {
    collapsed = !collapsed;
  }
</script>

<div class="container">
  <button class="flex gap-1" on:click={toggleCollapse}>
    <span class="header">{title}</span>
    <div class="transition-transform" class:-rotate-180={!collapsed}>
      <CaretDownIcon size="12px" />
    </div>
  </button>

  <div class="w-full h-fit max-h-full overflow-y-scroll px-[2px] pb-2">
    {#if !collapsed}
      {#if items.length}
        <DragList {items} />
      {:else}
        <p class="text-gray-500 my-1">No available fields</p>
      {/if}
    {/if}
  </div>
</div>

<style lang="postcss">
  .container {
    @apply pt-3 px-4;
    @apply flex flex-col gap-1 items-start;
    @apply h-fit max-h-fit min-h-8;
    @apply w-full min-w-60;
    @apply overflow-hidden;
    @apply flex-1;
  }

  button {
    @apply flex items-center justify-center;
  }

  .header {
    @apply uppercase font-semibold text-[10px];
  }
</style>
