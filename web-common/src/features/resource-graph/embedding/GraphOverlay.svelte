<script lang="ts">
  import type { ResourceGraphGrouping } from "../graph-canvas/graph-builder";
  import GraphCanvas from "../graph-canvas/GraphCanvas.svelte";
  import { FIT_VIEW_CONFIG } from "../shared/config";

  export let group: ResourceGraphGrouping | null = null;
  export let open = false;
  export let mode: "inline" | "fullscreen" | "modal" = "inline";
  export let showControls = true;
  export let showCloseButton = true;
  export let onClose: () => void;

  // Fit view configuration for better centering
  export let fitViewPadding: number = FIT_VIEW_CONFIG.PADDING;
  export let fitViewMinZoom: number = FIT_VIEW_CONFIG.MIN_ZOOM;
  export let fitViewMaxZoom: number = FIT_VIEW_CONFIG.MAX_ZOOM;

  function handleClose() {
    onClose();
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === "Escape") {
      handleClose();
    }
  }

  // Focus management for accessibility in fullscreen/modal modes
  let overlayEl: HTMLDivElement | null = null;
  let previouslyFocused: HTMLElement | null = null;

  $: if (open && overlayEl && (mode === "fullscreen" || mode === "modal")) {
    // Store the previously focused element before opening
    previouslyFocused = document.activeElement as HTMLElement;
    overlayEl.focus();
  } else if (!open && previouslyFocused) {
    // Restore focus when closing
    previouslyFocused.focus();
    previouslyFocused = null;
  }
</script>

{#if open && group}
  <!-- svelte-ignore a11y-no-noninteractive-tabindex -->
  <div
    bind:this={overlayEl}
    class="graph-overlay graph-overlay-{mode}"
    on:keydown={handleKeydown}
    role={mode === "inline" ? "region" : "dialog"}
    aria-modal={mode !== "inline"}
    aria-label="Expanded graph view"
    tabindex={mode !== "inline" ? 0 : undefined}
  >
    <div class="overlay-content">
      {#if showCloseButton && mode !== "inline"}
        <button
          class="close-btn"
          on:click={handleClose}
          aria-label="Close expanded graph"
          title="Close (ESC)"
        >
          ×
        </button>
      {/if}

      <GraphCanvas
        flowId={group.id}
        resources={group.resources}
        title={null}
        titleLabel={group.label}
        titleErrorCount={null}
        anchorError={false}
        rootNodeIds={undefined}
        fillParent
        {showControls}
        showLock={false}
        enableExpand={false}
        {fitViewPadding}
        {fitViewMinZoom}
        {fitViewMaxZoom}
      />
    </div>

    {#if mode !== "inline"}
      <!-- Backdrop for fullscreen/modal modes -->
      <div
        class="overlay-backdrop"
        on:click={handleClose}
        aria-hidden="true"
      ></div>
    {/if}
  </div>
{/if}

<style lang="postcss">
  .graph-overlay {
    @apply relative;
  }

  /* Inline mode: expands within the grid */
  .graph-overlay-inline {
    @apply col-span-full;
    height: 700px;
  }

  @media (min-width: 768px) {
    .graph-overlay-inline {
      height: 860px;
    }
  }

  /* Fullscreen mode: covers entire viewport */
  .graph-overlay-fullscreen {
    @apply fixed inset-0 z-50 bg-surface-background;
  }

  .graph-overlay-fullscreen .overlay-content {
    @apply relative h-full w-full p-4;
    z-index: 51;
  }

  .graph-overlay-fullscreen .overlay-backdrop {
    @apply hidden;
  }

  /* Modal mode: centered overlay with backdrop */
  .graph-overlay-modal {
    @apply fixed inset-0 z-50 flex items-center justify-center p-4;
  }

  .graph-overlay-modal .overlay-content {
    @apply relative h-full w-full max-w-7xl rounded-lg border border-gray-200 bg-surface-background p-4 shadow-xl;
    z-index: 51;
  }

  .graph-overlay-modal .overlay-backdrop {
    @apply fixed inset-0 bg-black/50;
    z-index: 50;
  }

  .close-btn {
    @apply absolute right-4 top-4 z-[52] flex h-8 w-8 items-center justify-center rounded-md border bg-surface-background text-2xl font-light text-fg-secondary;
    line-height: 1;
  }

  .close-btn:hover {
    @apply bg-muted text-fg-primary;
  }

  /* Ensure inline mode content fills available space */
  .graph-overlay-inline .overlay-content {
    @apply h-full w-full;
  }
</style>
