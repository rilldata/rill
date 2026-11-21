<script lang="ts">
  type Props = {
    title: string;
    colors: string[];
    min: string;
    max: string;
    /** Shown below the scale, e.g. to report dimension values that matched no region. */
    note?: string;
    noteTooltip?: string;
  };

  let { title, colors, min, max, note, noteTooltip }: Props = $props();

  const gradient = $derived(`linear-gradient(to right, ${colors.join(", ")})`);
</script>

<div class="legend">
  <div class="title" {title}>{title}</div>
  <div class="scale" style:background={gradient}></div>
  <div class="labels">
    <span>{min}</span>
    <span>{max}</span>
  </div>
  {#if note}
    <div class="note" title={noteTooltip}>{note}</div>
  {/if}
</div>

<style lang="postcss">
  .legend {
    @apply flex flex-col gap-1 w-40 px-2 py-1.5 rounded;
    @apply bg-surface-background/90 text-fg-secondary text-[11px] shadow-sm;
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
    @apply text-fg-muted truncate;
  }
</style>
