<script lang="ts">
  import { CanvasTocController } from "./toc-controller.svelte";

  type Props = {
    // The canvas scroll container (`#canvas-scroll-container`). The TOC is derived entirely from the
    // headings rendered inside its `.row-container`, so the component needs nothing else.
    scrollContainer: HTMLElement;
    // On narrow viewports the rail collapses its bars to small dots that tuck into the content's
    // left padding, so they don't overlap the content without reserving any layout width.
    compact?: boolean;
  };

  let { scrollContainer, compact = false }: Props = $props();

  let toc = $state<CanvasTocController>();
  $effect(() => {
    const controller = new CanvasTocController(scrollContainer);
    toc = controller;
    return () => controller.destroy();
  });
</script>

{#if toc && toc.entries.length > 1}
  <!-- Rest state shows a rail of tick bars; hovering or focusing into it reveals the full list.
       The same links back both states, so the accessible name is always the full heading text.
       Hidden when there's only one section, where a table of contents adds no value. -->
  <nav aria-label="Table of contents" class="toc" class:toc-compact={compact}>
    <div class="toc-panel">
      <ul class="toc-list">
        {#each toc.entries as entry (entry.id)}
          <li>
            <a
              href={`#${entry.id}`}
              class="toc-item"
              class:toc-item-active={toc.activeId === entry.id}
              style:--depth={entry.depth}
              aria-current={toc.activeId === entry.id ? "location" : undefined}
              onclick={(e) => toc?.jumpTo(e, entry)}
            >
              <span class="toc-bar" aria-hidden="true"></span>
              <span class="toc-text">{entry.text}</span>
            </a>
          </li>
        {/each}
      </ul>
    </div>
  </nav>
{/if}

<style lang="postcss">
  /* Full-height positioning anchor pinned to the left edge, out of flow, so the rail sits in the
     content's left margin without shifting the centered content or reserving layout width. It's
     transparent to pointer events (only `.toc-panel` is interactive), so hovering the empty gutter
     above/below the rail never holds the flyout open. */
  .toc {
    @apply pointer-events-none absolute left-0 top-0 z-40 h-full;
  }

  /* The rail itself: vertically centered, expanding to the full flyout on hover/focus (which
     overlays the content to its right rather than reflowing it). `p-2` keeps the first and last bars
     clear of the rounded corners so they aren't clipped. */
  .toc-panel {
    @apply pointer-events-auto absolute left-0 top-1/2 flex max-h-[calc(100%-1rem)] -translate-y-1/2 flex-col overflow-y-auto overflow-x-hidden;
    @apply rounded-lg border border-transparent p-2;
    @apply transition-[width,background-color,box-shadow,border-color] duration-150 motion-reduce:transition-none;
    width: 2rem;
  }

  .toc-panel:hover,
  .toc-panel:focus-within {
    @apply w-64 border-gray-200 bg-surface-card p-3 shadow-lg;
  }

  /* Rail: bars evenly spaced; rows stretch full width so the expanded hover highlight fills the row. */
  .toc-list {
    @apply flex flex-col gap-y-2;
  }

  .toc-panel:hover .toc-list,
  .toc-panel:focus-within .toc-list {
    @apply gap-y-0.5;
  }

  .toc-item {
    @apply flex w-full items-center rounded-md;
    @apply focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring;
  }

  /* Rail tick: darker and clearly sized, shrinking one step per nesting level so the hierarchy
     reads at a glance. Hidden once the panel expands (decorative only). */
  .toc-bar {
    @apply block h-[3px] flex-none rounded-full bg-gray-300;
    width: max(0.4rem, calc(1rem - var(--depth, 0) * 0.25rem));
  }

  .toc-item-active .toc-bar {
    @apply h-1 bg-primary-500;
  }

  .toc-panel:hover .toc-bar,
  .toc-panel:focus-within .toc-bar {
    @apply hidden;
  }

  /* Compact (narrow viewport): the collapsed rail shrinks to a column of small dots hugging the
     left edge so it tucks into the content's padding instead of overlapping it. The `:not(:hover)`
     guard keeps the hover/focus flyout expanding to the full panel as usual. */
  .toc-compact .toc-panel:not(:hover):not(:focus-within) {
    @apply w-4 px-1;
  }

  .toc-compact .toc-bar {
    @apply h-1 w-1 rounded-full;
  }

  .toc-compact .toc-item-active .toc-bar {
    @apply h-1.5 w-1.5;
  }

  /* Heading text: visually hidden but kept in the accessibility tree while in the rail (so each
     link's accessible name stays the full heading), then revealed as a normal list item. */
  .toc-text {
    @apply absolute h-px w-px overflow-hidden;
    clip: rect(0 0 0 0);
  }

  .toc-panel:hover .toc-text,
  .toc-panel:focus-within .toc-text {
    @apply static h-auto w-auto truncate;
    clip: auto;
  }

  /* Expanded list item: full-width row, indented per depth, muted by default, with a full-row hover
     highlight (base `.toc-item` already supplies `rounded-md`). */
  .toc-panel:hover .toc-item,
  .toc-panel:focus-within .toc-item {
    @apply py-1.5 pr-2 text-sm text-fg-secondary transition-colors;
    @apply hover:cursor-pointer hover:bg-surface-subtle hover:text-fg-primary;
    padding-left: calc(0.5rem + var(--depth, 0) * 0.875rem);
  }

  /* Active section: a soft tinted pill rather than a hard left bar. Text stays the normal
     foreground colour (white in dark mode) for readability; the tint conveys selection. */
  .toc-panel:hover .toc-item-active,
  .toc-panel:focus-within .toc-item-active {
    @apply bg-primary-50 font-medium text-fg-primary hover:bg-primary-100;
  }
</style>
