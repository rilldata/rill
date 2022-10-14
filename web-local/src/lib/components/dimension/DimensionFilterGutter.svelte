<script lang="ts">
  import { getContext } from "svelte";
  import StickyHeader from "../virtualized-table/core/StickyHeader.svelte";
  import type { VirtualizedTableConfig } from "../virtualized-table/types";
  import Check from "../icons/Check.svelte";
  import Cancel from "../icons/Cancel.svelte";
  import Spacer from "../icons/Spacer.svelte";

  const config: VirtualizedTableConfig = getContext("config");
  export let totalHeight: number;
  export let virtualRowItems;
  export let selectedIndex = [];
  export let excludeMode = false;
</script>

<div
  class="sticky left-0 top-0 z-20"
  style:height="{totalHeight}px"
  style:width="{config.indexWidth}px"
>
  {#each virtualRowItems as row (`row-${row.key}`)}
    {@const isSelected = selectedIndex.includes(row.index)}
    <StickyHeader
      enableResize={false}
      position="left"
      header={{ size: config.indexWidth, start: row.start }}
    >
      <div class="grid place-items-center">
        {#if isSelected && !excludeMode}
          <Check size="20px" />
        {:else if isSelected && excludeMode}
          <Cancel size="20px" />
        {:else}
          <Spacer />
        {/if}
      </div>
    </StickyHeader>
  {/each}
</div>
