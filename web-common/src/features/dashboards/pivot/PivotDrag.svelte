<script context="module" lang="ts">
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import DragList from "./DragList.svelte";
  import type { PivotSidebarSection, PivotChipData } from "./types";
</script>

<script lang="ts">
  export let title: PivotSidebarSection;
  export let items: PivotChipData[];
  export let collapsed = false;
  export let chipsPerSection: number;
  export let extraSpace: boolean;
  export let otherChipCounts: number[];

  $: fit =
    extraSpace ||
    items.length < chipsPerSection ||
    leavesSpaceForThirdSection();

  function leavesSpaceForThirdSection() {
    return otherChipCounts.some(
      (count) => count + items.length < chipsPerSection * 2,
    );
  }

  function toggleCollapse() {
    collapsed = !collapsed;
  }
</script>

<div class="container" class:fit class:full={!fit}>
  <button class="flex gap-1" on:click={toggleCollapse}>
    <span class="header">{title}</span>
    <div class="transition-transform" class:-rotate-180={!collapsed}>
      <CaretDownIcon size="12px" />
    </div>
  </button>

  <div class="w-full h-fit overflow-scroll px-[2px] pb-2">
    {#if !collapsed}
      {#if items.length}
        <DragList {items} type={title} />
      {:else}
        <p class="text-gray-500 my-1">No available fields</p>
      {/if}
    {/if}
  </div>
</div>

<style lang="postcss">
  .full {
    height: 100% !important;
    flex-shrink: 1 !important;
    /* This is enough to work in Safari without the JS workaround */
    /* flex-basis: 33% !important; */
    /* max-height: fit-content !important; */
  }

  .fit {
    height: fit-content !important;
    flex-shrink: 0 !important;
  }

  .container {
    @apply pt-3 px-4;
    @apply flex flex-col gap-1 items-start;
    @apply w-full overflow-hidden flex-grow-0;
    @apply border-b border-slate-200;
  }

  .container:last-of-type {
    @apply border-b-0;
  }

  button {
    @apply flex items-center justify-center;
  }

  .header {
    @apply uppercase font-semibold text-[10px];
  }
</style>
