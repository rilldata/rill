<script lang="ts">
  import { onDestroy, onMount } from "svelte";
  import { createScrollSpy } from "./scrollspy";
  import { deriveTocEntries, type TocEntry } from "./toc";

  // The canvas scroll container (`#canvas-scroll-container`). The TOC is derived entirely from the
  // headings rendered inside its `.row-container`, so the component needs nothing else.
  export let scrollContainer: HTMLElement;

  // On narrow viewports the rail collapses its bars to small dots that tuck into the content's left
  // padding, so they don't overlap the content without reserving any layout width.
  export let compact = false;

  let entries: TocEntry[] = [];

  let mutationObserver: MutationObserver | undefined;
  let debounceTimer: ReturnType<typeof setTimeout> | undefined;
  let deepLinked = false;
  let cancelScrollWatch: (() => void) | undefined;

  const spy = createScrollSpy(scrollContainer);
  const activeId = spy.activeId;

  onMount(() => {
    refresh();

    // Re-derive on any content change: tab switch, async markdown resolution, edits, filters, or
    // the `.row-container` itself appearing (e.g. after required filters are satisfied). Observing
    // the scroll container rather than `.row-container` keeps this resilient to remounts.
    mutationObserver = new MutationObserver(scheduleRefresh);
    mutationObserver.observe(scrollContainer, {
      childList: true,
      subtree: true,
      characterData: true,
    });
  });

  onDestroy(() => {
    clearTimeout(debounceTimer);
    cancelScrollWatch?.();
    mutationObserver?.disconnect();
    spy.destroy();
  });

  function scheduleRefresh() {
    clearTimeout(debounceTimer);
    debounceTimer = setTimeout(refresh, 150);
  }

  function refresh() {
    const rowContainer =
      scrollContainer.querySelector<HTMLElement>(".row-container");
    entries = deriveTocEntries(rowContainer);
    spy.setEntries(entries);

    // On first render, honor a deep link to a section (tab selection is handled separately via
    // the existing `tabs` URL param). Use an instant jump so the page doesn't animate on load.
    if (!deepLinked && entries.length > 0) {
      deepLinked = true;
      const hash = decodeURIComponent(window.location.hash.slice(1));
      const target = hash && entries.find((e) => e.id === hash);
      if (target) {
        target.el.scrollIntoView({ behavior: "auto", block: "start" });
        spy.setActive(target.id);
      }
    }
  }

  function scrollBehavior(): ScrollBehavior {
    return window.matchMedia?.("(prefers-reduced-motion: reduce)").matches
      ? "auto"
      : "smooth";
  }

  function jumpTo(event: MouseEvent, entry: TocEntry) {
    // Handle navigation ourselves so SvelteKit doesn't intercept the hash link, and so we can
    // respect prefers-reduced-motion.
    event.preventDefault();
    // Pin the highlight to the target for the duration of the scroll, so it doesn't flicker through
    // every section the smooth scroll passes; release once the scroll settles.
    spy.lock(entry.id);
    entry.el.scrollIntoView({ behavior: scrollBehavior(), block: "start" });
    history.replaceState(history.state, "", `#${entry.id}`);
    watchScrollEnd();
    // Release focus so the flyout collapses once the pointer leaves; otherwise `:focus-within`
    // keeps it open until the user clicks elsewhere.
    (event.currentTarget as HTMLElement).blur();
  }

  // Unlock the scrollspy once the container stops scrolling (debounced), with a hard timeout in case
  // no scroll events fire (e.g. the target is already in view, or reduced-motion jumps instantly).
  function watchScrollEnd() {
    cancelScrollWatch?.();

    let idle: ReturnType<typeof setTimeout>;
    const finish = () => {
      cancelScrollWatch?.();
      spy.unlock();
    };
    const onScroll = () => {
      clearTimeout(idle);
      idle = setTimeout(finish, 100);
    };
    const maxWait = setTimeout(finish, 1000);

    scrollContainer.addEventListener("scroll", onScroll, { passive: true });
    cancelScrollWatch = () => {
      clearTimeout(idle);
      clearTimeout(maxWait);
      scrollContainer.removeEventListener("scroll", onScroll);
      cancelScrollWatch = undefined;
    };
  }
</script>

{#if entries.length > 0}
  <!-- Rest state shows a rail of tick bars; hovering or focusing into it reveals the full list.
       The same links back both states, so the accessible name is always the full heading text. -->
  <nav aria-label="Table of contents" class="toc" class:toc--compact={compact}>
    <div class="toc__panel">
      <ul class="toc__list">
        {#each entries as entry (entry.id)}
          <li>
            <a
              href={`#${entry.id}`}
              class="toc__item"
              class:toc__item--active={$activeId === entry.id}
              style:--depth={entry.depth}
              aria-current={$activeId === entry.id ? "location" : undefined}
              on:click={(e) => jumpTo(e, entry)}
            >
              <span class="toc__bar" aria-hidden="true"></span>
              <span class="toc__text">{entry.text}</span>
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
     transparent to pointer events (only `.toc__panel` is interactive), so hovering the empty gutter
     above/below the rail never holds the flyout open. */
  .toc {
    @apply pointer-events-none absolute left-0 top-0 z-40 h-full;
  }

  /* The rail itself: vertically centered, expanding to the full flyout on hover/focus (which
     overlays the content to its right rather than reflowing it). `p-2` keeps the first and last bars
     clear of the rounded corners so they aren't clipped. */
  .toc__panel {
    @apply pointer-events-auto absolute left-0 top-1/2 flex max-h-[calc(100%-1rem)] -translate-y-1/2 flex-col overflow-y-auto overflow-x-hidden;
    @apply rounded-lg border border-transparent p-2;
    @apply transition-[width,background-color,box-shadow,border-color] duration-150 motion-reduce:transition-none;
    width: 2rem;
  }

  .toc__panel:hover,
  .toc__panel:focus-within {
    @apply w-64 border-gray-200 bg-surface-card p-3 shadow-lg;
  }

  /* Rail: bars evenly spaced; rows stretch full width so the expanded hover highlight fills the row. */
  .toc__list {
    @apply flex flex-col gap-y-2;
  }

  .toc__panel:hover .toc__list,
  .toc__panel:focus-within .toc__list {
    @apply gap-y-0.5;
  }

  .toc__item {
    @apply flex w-full items-center rounded-md;
    @apply focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring;
  }

  /* Rail tick: darker and clearly sized, shrinking one step per nesting level so the hierarchy
     reads at a glance. Hidden once the panel expands (decorative only). */
  .toc__bar {
    @apply block h-[3px] flex-none rounded-full bg-gray-300;
    width: max(0.4rem, calc(1rem - var(--depth, 0) * 0.25rem));
  }

  .toc__item--active .toc__bar {
    @apply h-1 bg-primary-500;
  }

  .toc__panel:hover .toc__bar,
  .toc__panel:focus-within .toc__bar {
    @apply hidden;
  }

  /* Compact (narrow viewport): the collapsed rail shrinks to a column of small dots hugging the
     left edge so it tucks into the content's padding instead of overlapping it. The `:not(:hover)`
     guard keeps the hover/focus flyout expanding to the full panel as usual. */
  .toc--compact .toc__panel:not(:hover):not(:focus-within) {
    @apply w-4 px-1;
  }

  .toc--compact .toc__bar {
    @apply h-1 w-1 rounded-full;
  }

  .toc--compact .toc__item--active .toc__bar {
    @apply h-1.5 w-1.5;
  }

  /* Heading text: visually hidden but kept in the accessibility tree while in the rail (so each
     link's accessible name stays the full heading), then revealed as a normal list item. */
  .toc__text {
    @apply absolute h-px w-px overflow-hidden;
    clip: rect(0 0 0 0);
  }

  .toc__panel:hover .toc__text,
  .toc__panel:focus-within .toc__text {
    @apply static h-auto w-auto truncate;
    clip: auto;
  }

  /* Expanded list item: full-width row, indented per depth, muted by default, with a full-row hover
     highlight (base `.toc__item` already supplies `rounded-md`). */
  .toc__panel:hover .toc__item,
  .toc__panel:focus-within .toc__item {
    @apply py-1.5 pr-2 text-sm text-fg-secondary transition-colors;
    @apply hover:cursor-pointer hover:bg-surface-subtle hover:text-fg-primary;
    padding-left: calc(0.5rem + var(--depth, 0) * 0.875rem);
  }

  /* Active section: a soft tinted pill rather than a hard left bar. Text stays the normal
     foreground colour (white in dark mode) for readability; the tint conveys selection. */
  .toc__panel:hover .toc__item--active,
  .toc__panel:focus-within .toc__item--active {
    @apply bg-primary-50 font-medium text-fg-primary hover:bg-primary-100;
  }
</style>
