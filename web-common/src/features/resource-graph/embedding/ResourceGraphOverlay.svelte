<script lang="ts">
  import Overlay from "@rilldata/web-common/components/overlay/Overlay.svelte";
  import { X, GitBranch } from "lucide-svelte";
  import ResourceGraph from "./ResourceGraph.svelte";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { ALLOWED_FOR_GRAPH } from "../navigation/seed-parser";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { tick } from "svelte";

  export let open = false;
  export let onClose: (() => void) | undefined = undefined;
  export let anchorResource: V1Resource | undefined;
  export let resources: V1Resource[] = [];
  export let isLoading = false;
  export let error: string | null = null;

  // Type for resource kinds that support graph visualization
  type GraphableKind =
    | ResourceKind.Source
    | ResourceKind.Model
    | ResourceKind.MetricsView
    | ResourceKind.Explore
    | ResourceKind.Canvas;

  const NAME_SEED_ALIAS: Record<GraphableKind, string> = {
    [ResourceKind.Source]: "model", // Sources are treated as models (deprecated)
    [ResourceKind.Model]: "model",
    [ResourceKind.MetricsView]: "metrics",
    [ResourceKind.Explore]: "dashboard",
    [ResourceKind.Canvas]: "dashboard",
  };

  const KIND_TOKEN_BY_KIND: Record<GraphableKind, string> = {
    [ResourceKind.Source]: "models", // Sources are treated as models (deprecated)
    [ResourceKind.Model]: "models",
    [ResourceKind.MetricsView]: "metrics",
    [ResourceKind.Explore]: "dashboards",
    [ResourceKind.Canvas]: "dashboards",
  };

  $: anchorName = anchorResource?.meta?.name?.name ?? null;
  $: anchorKind = anchorResource?.meta?.name?.kind as ResourceKind | undefined;
  // Check if kind is allowed (handles both enum and string values)
  // Convert to string for comparison since ResourceKind enum values are strings
  $: supportsGraph = anchorKind ? ALLOWED_FOR_GRAPH.has(String(anchorKind) as ResourceKind) : false;

  // Type-safe access to graphable kind properties
  $: graphableKind =
    supportsGraph && anchorKind ? (anchorKind as GraphableKind) : null;

  // Use the same seed format as the project graph page
  // This ensures consistent behavior between overlay and project graph
  $: overlaySeeds = (function (): string[] | undefined {
    if (!anchorName || !anchorKind) return undefined;
    
    // Normalize kind to string for comparison (handles both enum and string values)
    const kindStr = String(anchorKind);
    
    // Use the same format that would come from URL parameters
    // For Canvas/Explore, use "dashboard:" prefix (same as project graph)
    // For other resources, use their kind prefix
    if (kindStr === ResourceKind.Canvas || kindStr === ResourceKind.Explore) {
      return [`dashboard:${anchorName}`];
    } else if (kindStr === ResourceKind.Source || kindStr === ResourceKind.Model) {
      return [`model:${anchorName}`];
    } else if (kindStr === ResourceKind.MetricsView) {
      return [`metrics:${anchorName}`];
    }
    
    // Fallback to kind:name format using the resource's actual kind
    const kindPart = kindStr.toLowerCase().replace("rill.runtime.v1.", "").replace("metricsview", "metrics");
    return [`${kindPart}:${anchorName}`];
  })();
  
  // Keep the old anchorSeed for graphHref calculation
  $: anchorSeed =
    graphableKind && anchorName
      ? `${NAME_SEED_ALIAS[graphableKind]}:${anchorName}`
      : null;
  $: graphHref = graphableKind
    ? `/graph?kind=${encodeURIComponent(KIND_TOKEN_BY_KIND[graphableKind])}`
    : "/graph";

  $: emptyReason = !overlaySeeds || overlaySeeds.length === 0 ? "unsupported" : null;

  function closeOverlay() {
    if (onClose) {
      onClose();
    } else {
      // Fallback for bind:open usage
      open = false;
    }
  }

  function handleDialogClick(event: MouseEvent) {
    // Prevent clicks inside the dialog from bubbling to the overlay.
    event.stopPropagation();
  }
  let backdropInteractive = false;
  $: if (open) {
    backdropInteractive = false;
    tick().then(() => {
      if (open) backdropInteractive = true;
    });
  }

  function handleBackdropClick() {
    if (!backdropInteractive) return;
    closeOverlay();
  }
</script>

{#if open}
  <Overlay bg="rgba(15,23,42,0.8)">
    <!-- svelte-ignore a11y-click-events-have-key-events -->
    <!-- svelte-ignore a11y-no-static-element-interactions -->
    <div
      class="graph-overlay__backdrop"
      on:click={handleBackdropClick}
      role="presentation"
    >
      <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
      <div
        class="graph-overlay"
        role="dialog"
        aria-modal="true"
        aria-label={anchorName
          ? `Resource graph for ${anchorName}`
          : "Resource graph"}
        on:click={handleDialogClick}
      >
        <header class="graph-overlay__header">
          <div class="graph-overlay__title">
            <GitBranch size="16px" aria-hidden="true" />
            <div>
              <p class="graph-overlay__eyebrow">Resource graph</p>
              <h2>{anchorName ?? "Select a resource"}</h2>
            </div>
          </div>
          <div class="graph-overlay__actions">
            <a class="graph-overlay__link" href={graphHref} rel="noreferrer">
              Project Graphs
            </a>
            <button
              class="graph-overlay__close"
              on:click={closeOverlay}
              aria-label="Close resource graph overlay"
            >
              <X size="18px" aria-hidden="true" />
            </button>
          </div>
        </header>

        <section class="graph-overlay__body">
          {#if error}
            <p class="graph-overlay__state graph-overlay__error">
              {error}
            </p>
          {:else if emptyReason === "unsupported"}
            <p class="graph-overlay__state">
              This resource type doesn't have a project graph view.
            </p>
          {:else}
            <div class="graph-overlay__graph">
              <ResourceGraph
                {resources}
                {isLoading}
                {error}
                seeds={overlaySeeds}
                syncExpandedParam={false}
                showSummary={false}
                showCardTitles={false}
                maxGroups={null}
                showControls={false}
                enableExpansion={false}
                fitViewPadding={0.08}
                fitViewMinZoom={0.01}
                fitViewMaxZoom={1.35}
                expandedHeightMobile="100%"
                expandedHeightDesktop="100%"
              />
            </div>
          {/if}
        </section>
      </div>
    </div>
  </Overlay>
{/if}

<style lang="postcss">
  .graph-overlay__backdrop {
    @apply fixed inset-0 flex items-center justify-center w-full h-full;
  }

  .graph-overlay {
    @apply bg-surface border border-gray-200 rounded-xl shadow-2xl overflow-hidden;
    @apply flex flex-col;
    width: min(1100px, 90vw);
    height: min(80vh, 760px);
  }

  .graph-overlay__header {
    @apply flex items-start justify-between gap-x-4 px-5 py-4 border-b border-gray-200;
  }

  .graph-overlay__title {
    @apply flex items-center gap-x-3 text-left;
  }

  .graph-overlay__eyebrow {
    @apply text-xs uppercase text-gray-500 tracking-wide;
  }

  .graph-overlay__title h2 {
    @apply text-lg font-semibold text-gray-900 leading-snug;
  }

  .graph-overlay__actions {
    @apply flex items-center gap-x-2;
  }

  .graph-overlay__link {
    @apply text-xs font-medium text-primary-700 border border-primary-200 rounded-full px-3 py-1 transition-colors;
  }

  .graph-overlay__link:hover {
    @apply bg-primary-50;
  }

  .graph-overlay__close {
    @apply rounded-full border border-gray-200 text-gray-500 p-1;
  }

  .graph-overlay__close:hover {
    @apply text-gray-700 border-gray-300;
  }

  .graph-overlay__body {
    @apply flex-1 flex flex-col w-full bg-gray-50 min-h-0;
  }

  .graph-overlay__graph {
    @apply flex-1 min-h-0;
  }

  .graph-overlay__state {
    @apply text-sm text-gray-600 m-auto text-center max-w-sm;
  }

  .graph-overlay__error {
    @apply text-red-600;
  }
</style>
