<script lang="ts">
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import { GitBranch } from "lucide-svelte";
  import ResourceGraph from "./ResourceGraph.svelte";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { ALLOWED_FOR_GRAPH } from "../navigation/seed-parser";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";

  export let open = false;
  export let onClose: (() => void) | undefined = undefined;
  export let anchorResource: V1Resource | undefined;
  export let resources: V1Resource[] = [];
  export let isLoading = false;
  export let error: string | null = null;

  type GraphableKind =
    | ResourceKind.Source
    | ResourceKind.Model
    | ResourceKind.MetricsView
    | ResourceKind.Explore
    | ResourceKind.Canvas;

  const KIND_TOKEN_BY_KIND: Record<GraphableKind, string> = {
    [ResourceKind.Source]: "sources",
    [ResourceKind.Model]: "models",
    [ResourceKind.MetricsView]: "metrics",
    [ResourceKind.Explore]: "dashboards",
    [ResourceKind.Canvas]: "dashboards",
  };

  $: anchorName = anchorResource?.meta?.name?.name ?? null;
  $: rawAnchorKind = anchorResource?.meta?.name?.kind as
    | ResourceKind
    | undefined;
  $: anchorKind =
    rawAnchorKind === ResourceKind.Source ? ResourceKind.Model : rawAnchorKind;

  $: supportsGraph = anchorKind
    ? ALLOWED_FOR_GRAPH.has(String(anchorKind) as ResourceKind)
    : false;

  $: graphableKind =
    supportsGraph && anchorKind ? (anchorKind as GraphableKind) : null;

  $: overlaySeeds = (function (): string[] | undefined {
    if (!anchorName || !anchorKind) return undefined;
    if (anchorKind === ResourceKind.Canvas) {
      return [`canvas:${anchorName}`];
    } else if (anchorKind === ResourceKind.Explore) {
      return [`explore:${anchorName}`];
    } else if (anchorKind === ResourceKind.Model) {
      return [`model:${anchorName}`];
    } else if (anchorKind === ResourceKind.MetricsView) {
      return [`metrics:${anchorName}`];
    }
    return undefined;
  })();

  $: graphHref = graphableKind
    ? `/graph?kind=${encodeURIComponent(KIND_TOKEN_BY_KIND[graphableKind])}`
    : "/graph";

  $: emptyReason =
    !overlaySeeds || overlaySeeds.length === 0 ? "unsupported" : null;

  function handleOpenChange(isOpen: boolean) {
    if (!isOpen) {
      onClose?.();
      open = false;
    }
  }
</script>

<Dialog.Root {open} onOpenChange={handleOpenChange}>
  <Dialog.Content class="graph-dialog">
    <div class="graph-dialog__header">
      <div class="graph-dialog__title">
        <GitBranch size="16px" aria-hidden="true" />
        <div>
          <p class="graph-dialog__eyebrow">Resource graph</p>
          <Dialog.Title class="graph-dialog__name">
            {anchorName ?? "Select a resource"}
          </Dialog.Title>
        </div>
      </div>
      <a
        class="graph-dialog__link"
        href={graphHref}
        rel="noreferrer"
        on:click={handleOpenChange.bind(null, false)}
      >
        Project Graphs
      </a>
    </div>

    <div class="graph-dialog__body">
      {#if error}
        <p class="graph-dialog__state graph-dialog__error">{error}</p>
      {:else if emptyReason === "unsupported"}
        <p class="graph-dialog__state">
          This resource type doesn't have a project graph view.
        </p>
      {:else}
        <div class="graph-dialog__graph">
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
            showNodeActions={false}
            enableExpansion={false}
            fitViewPadding={0.08}
            fitViewMinZoom={0.01}
            fitViewMaxZoom={1.35}
            expandedHeightMobile="100%"
            expandedHeightDesktop="100%"
          />
        </div>
      {/if}
    </div>
  </Dialog.Content>
</Dialog.Root>

<style lang="postcss">
  :global(.graph-dialog) {
    display: flex !important;
    flex-direction: column !important;
    width: min(1100px, 90vw) !important;
    max-width: min(1100px, 90vw) !important;
    height: min(80vh, 760px) !important;
    padding: 0 !important;
    gap: 0 !important;
    overflow: hidden !important;
  }

  .graph-dialog__header {
    @apply flex items-start justify-between gap-x-4 px-5 py-4 pr-12 border-b flex-none;
  }

  .graph-dialog__title {
    @apply flex items-center gap-x-3 text-left;
  }

  .graph-dialog__eyebrow {
    @apply text-xs uppercase text-fg-secondary tracking-wide;
  }

  :global(.graph-dialog__name) {
    @apply text-lg font-semibold text-fg-primary leading-snug;
  }

  .graph-dialog__link {
    @apply text-xs font-medium text-primary-700 border border-primary-200 rounded-full px-3 py-1 transition-colors;
  }

  .graph-dialog__link:hover {
    @apply bg-primary-50;
  }

  .graph-dialog__body {
    @apply flex-1 flex flex-col w-full bg-surface-background min-h-0;
  }

  .graph-dialog__graph {
    @apply flex-1 min-h-0;
  }

  .graph-dialog__state {
    @apply text-sm text-fg-secondary m-auto text-center max-w-sm;
  }

  .graph-dialog__error {
    @apply text-red-600;
  }
</style>
