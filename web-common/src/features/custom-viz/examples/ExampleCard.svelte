<script lang="ts">
  import ChartIcon from "@rilldata/web-common/features/canvas/icons/ChartIcon.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import type { GalleryEntry } from "./example-spec";
  import { thumbnailURL } from "./thumbnails";

  export let example: GalleryEntry;
  export let onSelect: (example: GalleryEntry) => void;
  export let disabled = false;

  // Some upstream descriptions contain markdown links; show just the link text.
  $: description = example.description?.replace(/\[([^\]]+)\]\([^)]*\)/g, "$1");
  $: thumbnail = thumbnailURL(example);
</script>

<button
  class="flex flex-col text-left border rounded-md overflow-hidden hover:border-primary-400 hover:shadow-sm bg-surface-card disabled:opacity-50 disabled:pointer-events-none"
  onclick={() => onSelect(example)}
  {disabled}
  title={description ?? example.title}
>
  <!-- The Vega editor's own preview image, vendored at snapshot time: no chart
       engine runs at gallery time, so cards render instantly. -->
  <div class="thumbnail pointer-events-none" aria-hidden="true">
    {#if thumbnail}
      <img src={thumbnail} alt="" loading="lazy" />
    {:else}
      <span class="text-fg-secondary"><ChartIcon /></span>
    {/if}
  </div>
  <div class="px-3 py-2 border-t w-full">
    <div class="text-xs font-medium truncate">{example.title}</div>
    <div
      class="text-[11px] text-fg-secondary truncate flex items-center gap-x-1"
    >
      <span class="truncate">{example.subcategory}</span>
      {#if !example.parameterizes}
        <!-- This example can't be converted to a parameterized component
             deterministically; it imports with the sample data and has to be
             connected to project data afterwards. -->
        <span class="badge shrink-0">{m.component_example_sample_badge()}</span>
      {/if}
    </div>
  </div>
</button>

<style lang="postcss">
  .thumbnail {
    @apply h-32 w-full p-2 overflow-hidden bg-white flex items-center justify-center;
  }
  /* Upstream images vary widely in aspect ratio (tall bar charts, wide
     heatmaps); contain them so every card reads at the same size. */
  .thumbnail img {
    @apply max-h-full max-w-full object-contain;
  }
  .badge {
    @apply px-1 py-px rounded-sm bg-surface-subtle text-fg-secondary text-[10px] leading-tight;
  }
</style>
