<script lang="ts">
  import Overlay from "@rilldata/web-common/components/overlay/Overlay.svelte";
  import { X, GitBranch } from "lucide-svelte";
  import ResourceGraph from "./ResourceGraph.svelte";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { ALLOWED_FOR_GRAPH } from "./seed-utils";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { tick } from "svelte";

  export let open = false;
  export let anchorResource: V1Resource | undefined;
  export let resources: V1Resource[] = [];
  export let isLoading = false;
  export let error: string | null = null;

  const NAME_SEED_ALIAS: Partial<Record<ResourceKind, string>> = {
    [ResourceKind.Source]: "source",
    [ResourceKind.Model]: "model",
    [ResourceKind.MetricsView]: "metrics",
    [ResourceKind.Explore]: "dashboard",
  };
  const KIND_TOKEN_BY_KIND: Partial<Record<ResourceKind, string>> = {
    [ResourceKind.Source]: "sources",
    [ResourceKind.Model]: "models",
    [ResourceKind.MetricsView]: "metrics",
    [ResourceKind.Explore]: "dashboards",
  };

  $: anchorMeta = anchorResource?.meta;
  $: anchorName = anchorMeta?.name?.name ?? null;
  $: anchorKind = anchorMeta?.name?.kind as ResourceKind | undefined;
  $: supportsGraph = anchorKind
    ? ALLOWED_FOR_GRAPH.has(anchorKind)
    : false;
  function buildSeed(kind?: ResourceKind, name?: string | null) {
    if (!kind || !name) return null;
    const alias = NAME_SEED_ALIAS[kind] ?? kind;
    return `${alias}:${name}`;
  }

  $: anchorSeed = supportsGraph
    ? buildSeed(anchorKind, anchorMeta?.name?.name)
    : null;
  $: overlaySeeds = anchorSeed ? [anchorSeed] : undefined;
  $: graphHrefSeed = supportsGraph ? KIND_TOKEN_BY_KIND[anchorKind!] : null;
  $: graphHref = graphHrefSeed
    ? `/graph?seed=${encodeURIComponent(graphHrefSeed)}`
    : "/graph";

  $: emptyReason = !anchorSeed ? "unsupported" : null;

  function closeOverlay() {
    open = false;
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
    <div class="graph-overlay__backdrop" on:click={handleBackdropClick}>
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
        {#if emptyReason === "unsupported"}
          <p class="graph-overlay__state">
            This resource type doesn't have a project graph view.
          </p>
        {:else}
          <div class="graph-overlay__graph">
            <ResourceGraph
              resources={resources}
              {isLoading}
              {error}
              seeds={overlaySeeds}
              syncExpandedParam={false}
              showSummary={false}
              showCardTitles={false}
              maxGroups={1}
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
    @apply text-xs font-medium text-primary-700 border border-primary-200 rounded-full px-3 py-1 hover:bg-primary-50 transition-colors;
  }

  .graph-overlay__close {
    @apply rounded-full border border-gray-200 text-gray-500 hover:text-gray-700 hover:border-gray-300 p-1;
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
</style>
