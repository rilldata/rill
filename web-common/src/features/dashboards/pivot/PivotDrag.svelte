<script context="module" lang="ts">
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import DragList from "./DragList.svelte";
  import type { PivotSidebarSection, PivotChipData } from "./types";
  import { afterUpdate } from "svelte";
</script>

<script lang="ts">
  export let title: PivotSidebarSection;
  export let items: PivotChipData[];
  export let collapsed = false;

  let container: HTMLDivElement;

  afterUpdate(() => {
    if (!container) return;
    console.log("bottom");
    calculateSize(container);
  });

  function toggleCollapse() {
    collapsed = !collapsed;
  }

  // Only Safari seems to support flex-basis in conjunction with max-height: fit-content
  // So, this is a workaround to achieve the same thing in JS
  function calculateSize(element: HTMLDivElement) {
    element.style.height = "fit-content";
    element.style.flexShrink = "0";

    const fitContentHeight = container.offsetHeight;

    element.style.height = "100%";
    element.style.flexShrink = "1";

    const evenSplitHeight = container.offsetHeight;

    if (fitContentHeight < evenSplitHeight) {
      element.style.height = "fit-content";
      element.style.flexShrink = "0";
    }
  }
</script>

<svelte:window on:resize={() => calculateSize(container)} />

<div class="container" bind:this={container}>
  <button class="flex gap-1" on:click={toggleCollapse}>
    <span class="header">{title}</span>
    <div class="transition-transform" class:-rotate-180={!collapsed}>
      <CaretDownIcon size="12px" />
    </div>
  </button>

  <div class="w-full h-fit overflow-scroll px-[2px] pb-2">
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
    @apply w-full overflow-hidden flex-grow-0;
    @apply border-b border-slate-200;
  }

  .container:last-child {
    @apply border-b-0;
  }

  button {
    @apply flex items-center justify-center;
  }

  .header {
    @apply uppercase font-semibold text-[10px];
  }
</style>
