<script lang="ts">
  import ChartIcon from "@rilldata/web-common/features/canvas/icons/ChartIcon.svelte";
  import type { GalleryEntry } from "./example-spec";

  export let example: GalleryEntry;
  export let onSelect: (example: GalleryEntry) => void;
  export let disabled = false;
</script>

<button
  class="flex flex-col text-left border rounded-md overflow-hidden hover:border-primary-400 hover:shadow-sm bg-surface-card disabled:opacity-50 disabled:pointer-events-none"
  onclick={() => onSelect(example)}
  {disabled}
  title="{example.chartType} — {example.description || example.title}"
>
  <!-- Pre-rendered at generation time and inlined as a data URI, so browsing the gallery runs no
       chart engine and fetches no assets. -->
  <div class="thumbnail pointer-events-none" aria-hidden="true">
    {#if example.thumbnail}
      <img src={example.thumbnail} alt="" loading="lazy" />
    {:else}
      <span class="text-fg-secondary"><ChartIcon /></span>
    {/if}
  </div>
  <div class="px-3 py-2 border-t w-full">
    <div class="text-xs font-medium truncate">{example.chartType}</div>
    {#if example.description}
      <div class="text-[11px] text-fg-secondary truncate">
        {example.description}
      </div>
    {/if}
  </div>
</button>

<style lang="postcss">
  .thumbnail {
    @apply h-32 w-full p-2 overflow-hidden bg-white flex items-center justify-center;
  }
  /* Examples vary widely in aspect ratio (tall bar charts, wide heatmaps);
     contain them so every card reads at the same size. */
  .thumbnail img {
    @apply max-h-full max-w-full object-contain;
  }
</style>
