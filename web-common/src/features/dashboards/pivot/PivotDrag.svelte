<script context="module" lang="ts">
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import DragList from "./DragList.svelte";
  import type { PivotSidebarSection, PivotChipData } from "./types";
</script>

<script lang="ts">
  export let title: PivotSidebarSection;
  export let items: PivotChipData[];
  export let collapsed = false;
  export let disabled = false;

  let showMore = false;

  $: visible = showMore ? items.length : 3;

  function toggleCollapse() {
    collapsed = !collapsed;
  }

  function toggleShowMore() {
    showMore = !showMore;
  }
</script>

<div class="flex flex-col gap-1 items-start">
  <button class="flex gap-1" on:click={toggleCollapse}>
    <span class="header">{title}</span>
    <div class="transition-transform" class:-rotate-180={!collapsed}>
      <CaretDownIcon size="12px" />
    </div>
  </button>

  {#if !collapsed}
    <DragList {disabled} items={items.slice(0, visible)} />

    {#if !collapsed && items.length > 3}
      <button class="see-more" on:click={toggleShowMore}>
        {showMore ? "Show less" : "Show more"}
      </button>
    {/if}
  {/if}
</div>

<style lang="postcss">
  button {
    @apply flex items-center justify-center;
  }

  .see-more {
    @apply ml-2;
  }

  .header {
    @apply uppercase font-semibold text-[10px];
  }
</style>
