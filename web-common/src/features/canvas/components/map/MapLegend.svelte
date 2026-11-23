<script lang="ts">
  import type { MapLegendNote } from "./map-utils";

  type Props = {
    title: string;
    /** Names the layer the scale belongs to, when several layers show a legend. */
    layer?: string;
    colors: string[];
    min: string;
    max: string;
    /** Shown below the scale, e.g. to report dimension values that matched no region. */
    notes?: MapLegendNote[];
  };

  let { title, layer, colors, min, max, notes = [] }: Props = $props();

  const gradient = $derived(`linear-gradient(to right, ${colors.join(", ")})`);
</script>

<div class="legend">
  {#if layer}
    <div class="layer" title={layer}>{layer}</div>
  {/if}
  <div class="title" {title}>{title}</div>
  <div class="scale" style:background={gradient}></div>
  <div class="labels">
    <span>{min}</span>
    <span>{max}</span>
  </div>
  {#each notes as note (note.text)}
    <div class="note" title={note.tooltip}>
      {#if note.color}
        <span class="swatch" style:background={note.color}></span>
      {/if}
      <span class="note-text">{note.text}</span>
    </div>
  {/each}
</div>

<style lang="postcss">
  .legend {
    @apply flex flex-col gap-1 w-40 px-2 py-1.5 rounded;
    @apply bg-surface-background/90 text-fg-secondary text-[11px] shadow-sm;
  }

  .layer {
    @apply text-fg-muted truncate;
  }

  .title {
    @apply font-medium text-fg-primary truncate;
  }

  .scale {
    @apply h-2 rounded-sm;
  }

  .labels {
    @apply flex justify-between gap-2;
  }

  .note {
    @apply flex items-center gap-1 text-fg-muted;
  }

  .note-text {
    @apply truncate;
  }

  .swatch {
    @apply shrink-0 size-2 rounded-sm;
  }
</style>
