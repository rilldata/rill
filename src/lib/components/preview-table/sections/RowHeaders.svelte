<script lang="ts">
  import { config } from "../config";
  import StickyHeader from "../core/StickyHeader.svelte";
  export let totalHeight: number;
  export let virtualRowItems;
</script>

<div
  class="sticky left-0 top-0 z-20"
  style:height="{totalHeight}px"
  style:width="{config.indexWidth}px"
>
  <StickyHeader
    header={{ size: config.indexWidth, start: 0 }}
    enableResize={false}
    position="top-left">#</StickyHeader
  >
  {#each virtualRowItems as row (`row-${row.key}`)}
    <StickyHeader
      enableResize={false}
      position="left"
      header={{ size: config.indexWidth, start: row.start }}
    >
      {row.key + 1}
    </StickyHeader>
  {/each}
</div>
