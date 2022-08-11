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
    position="top-left">#</StickyHeader
  >
  {#each virtualRowItems as row (`row-${row.key}`)}
    <div
      class="absolute left-0 z-20 bg-gray-100 grid place-items-center font-bold border-r border-gray-300 border-b"
      style:height="{row.size}px"
      style:width="{config.indexWidth}px"
      style:left={0}
      style:top={0}
      style:transform="translateY({row.start}px)"
    >
      {row.key + 1}
    </div>
  {/each}
</div>
