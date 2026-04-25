<script lang="ts">
  import { page } from "$app/stores";
  import CanvasIcon from "@rilldata/web-common/components/icons/CanvasIcon.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import ExploreIcon from "@rilldata/web-common/components/icons/ExploreIcon.svelte";
  import MetricsViewIcon from "@rilldata/web-common/components/icons/MetricsViewIcon.svelte";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import {
    createRuntimeServiceListResources,
    type V1Resource,
  } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";

  type IconComponent =
    | typeof ExploreIcon
    | typeof CanvasIcon
    | typeof MetricsViewIcon;

  const runtimeClient = useRuntimeClient();

  $: exploresQuery = createRuntimeServiceListResources(
    runtimeClient,
    { kind: ResourceKind.Explore },
    { query: { select: (d) => d.resources ?? [] } },
  );
  $: canvasesQuery = createRuntimeServiceListResources(
    runtimeClient,
    { kind: ResourceKind.Canvas },
    { query: { select: (d) => d.resources ?? [] } },
  );
  $: metricsQuery = createRuntimeServiceListResources(
    runtimeClient,
    { kind: ResourceKind.MetricsView },
    { query: { select: (d) => d.resources ?? [] } },
  );

  $: explores = $exploresQuery.data ?? [];
  $: canvases = $canvasesQuery.data ?? [];
  $: metrics = $metricsQuery.data ?? [];

  $: currentFile = $page.params.file ? `/${$page.params.file}` : undefined;
  $: currentPath = $page.url.pathname;

  function fileHref(filePath: string | undefined): string | undefined {
    if (!filePath) return undefined;
    return `/files${filePath.startsWith("/") ? "" : "/"}${filePath}`;
  }

  type Item = {
    name: string;
    href: string;
    icon: IconComponent;
    activeWhenFile?: string;
    activeWhenPath?: string;
  };

  // A "synthetic" Explore is one the runtime auto-creates from a legacy v0
  // metrics view's embedded dashboard config. It has no explore YAML of its
  // own, so meta.filePaths[0] points back at the metrics view file. Detect
  // by file-path overlap.
  function isSyntheticExplore(
    explore: V1Resource,
    metricsFilePaths: Set<string>,
  ): boolean {
    const filePath = explore.meta?.filePaths?.[0];
    return !!filePath && metricsFilePaths.has(filePath);
  }

  // Real explores only — points the user at the editable dashboard YAML.
  // Synthetic explores are skipped here (would route the user back into the
  // metrics view file). Orphan v0 metrics views are picked up by
  // legacyMetricsAsDashboards instead.
  function exploreItems(
    explores: V1Resource[],
    metricsFilePaths: Set<string>,
  ): Item[] {
    return explores
      .filter((r) => !isSyntheticExplore(r, metricsFilePaths))
      .map((r): Item | null => {
        const name = r.meta?.name?.name ?? "";
        const filePath = r.meta?.filePaths?.[0];
        const href = fileHref(filePath);
        if (!name || !filePath || !href) return null;
        return { name, href, icon: ExploreIcon, activeWhenFile: filePath };
      })
      .filter((i): i is Item => i !== null);
  }

  function canvasItems(resources: V1Resource[]): Item[] {
    return resources
      .map((r): Item | null => {
        const name = r.meta?.name?.name ?? "";
        const filePath = r.meta?.filePaths?.[0];
        const href = fileHref(filePath);
        if (!name || !href) return null;
        return { name, href, icon: CanvasIcon, activeWhenFile: filePath };
      })
      .filter((i): i is Item => i !== null);
  }

  // A "legacy" metrics view here is one with no real Explore resource
  // referencing it. The single YAML file plays both roles, so we surface it
  // in the Dashboards section linked to /explore/<name>, the preview route
  // that already knows how to render legacy v0 dashboards. Synthetic
  // explores are ignored when computing the referenced set, since they
  // exist only because of the metrics view itself.
  function legacyMetricsAsDashboards(
    metrics: V1Resource[],
    explores: V1Resource[],
    metricsFilePaths: Set<string>,
  ): Item[] {
    const referencedByReal = new Set(
      explores
        .filter((e) => !isSyntheticExplore(e, metricsFilePaths))
        .map((e) => e.explore?.spec?.metricsView)
        .filter((n): n is string => !!n),
    );
    return metrics
      .filter((m) => {
        const name = m.meta?.name?.name;
        return name && !referencedByReal.has(name);
      })
      .map((m) => {
        const name = m.meta?.name?.name ?? "";
        const href = `/explore/${name}`;
        return {
          name,
          href,
          icon: ExploreIcon,
          activeWhenPath: href,
        };
      });
  }

  function metricsItems(resources: V1Resource[]): Item[] {
    return resources
      .map((r): Item | null => {
        const name = r.meta?.name?.name ?? "";
        const filePath = r.meta?.filePaths?.[0];
        const href = fileHref(filePath);
        if (!name || !href) return null;
        return { name, href, icon: MetricsViewIcon, activeWhenFile: filePath };
      })
      .filter((i): i is Item => i !== null);
  }

  function sortByName(items: Item[]): Item[] {
    return items.sort((a, b) =>
      a.name.localeCompare(b.name, undefined, { sensitivity: "base" }),
    );
  }

  $: metricsFilePaths = new Set(
    metrics.map((m) => m.meta?.filePaths?.[0]).filter((p): p is string => !!p),
  );

  $: dashboardItems = sortByName([
    ...exploreItems(explores, metricsFilePaths),
    ...legacyMetricsAsDashboards(metrics, explores, metricsFilePaths),
  ]);
  $: canvasNavItems = sortByName(canvasItems(canvases));
  $: metricsNavItems = sortByName(metricsItems(metrics));

  function isActive(item: Item): boolean {
    if (item.activeWhenPath) return currentPath === item.activeWhenPath;
    if (item.activeWhenFile) return currentFile === item.activeWhenFile;
    return false;
  }

  let dashboardsOpen = true;
  let canvasesOpen = true;
  let metricsOpen = true;
</script>

<nav class="flex flex-col gap-y-1 p-2 pb-6 w-full">
  {#if dashboardItems.length}
    <section>
      <button
        class="section-header"
        aria-expanded={dashboardsOpen}
        onclick={() => (dashboardsOpen = !dashboardsOpen)}
      >
        <CaretDownIcon
          size="16px"
          className="text-fg-secondary transition-transform {!dashboardsOpen &&
            '-rotate-90'}"
        />
        <h3>Dashboards</h3>
      </button>
      {#if dashboardsOpen}
        <ul>
          {#each dashboardItems as item (item.href)}
            <li>
              <a href={item.href} class="row" class:active={isActive(item)}>
                <span class="icon-wrap">
                  <svelte:component this={item.icon} size="22px" />
                </span>
                <span class="truncate">{item.name}</span>
              </a>
            </li>
          {/each}
        </ul>
      {/if}
    </section>
  {/if}

  {#if canvasNavItems.length}
    <section>
      <button
        class="section-header"
        aria-expanded={canvasesOpen}
        onclick={() => (canvasesOpen = !canvasesOpen)}
      >
        <CaretDownIcon
          size="16px"
          className="text-fg-secondary transition-transform {!canvasesOpen &&
            '-rotate-90'}"
        />
        <h3>Canvases</h3>
      </button>
      {#if canvasesOpen}
        <ul>
          {#each canvasNavItems as item (item.href)}
            <li>
              <a href={item.href} class="row" class:active={isActive(item)}>
                <span class="icon-wrap">
                  <svelte:component this={item.icon} size="22px" />
                </span>
                <span class="truncate">{item.name}</span>
              </a>
            </li>
          {/each}
        </ul>
      {/if}
    </section>
  {/if}

  {#if metricsNavItems.length}
    <section>
      <button
        class="section-header"
        aria-expanded={metricsOpen}
        onclick={() => (metricsOpen = !metricsOpen)}
      >
        <CaretDownIcon
          size="16px"
          className="text-fg-secondary transition-transform {!metricsOpen &&
            '-rotate-90'}"
        />
        <h3>Metrics</h3>
      </button>
      {#if metricsOpen}
        <ul>
          {#each metricsNavItems as item (item.href)}
            <li>
              <a href={item.href} class="row" class:active={isActive(item)}>
                <span class="icon-wrap">
                  <svelte:component this={item.icon} size="22px" />
                </span>
                <span class="truncate">{item.name}</span>
              </a>
            </li>
          {/each}
        </ul>
      {/if}
    </section>
  {/if}

  {#if !dashboardItems.length && !canvasNavItems.length && !metricsNavItems.length}
    <div class="empty-state">
      <p>No dashboards, canvases, or metrics views yet.</p>
      <p class="hint">
        Switch to <span class="mono">Code</span> mode to create them.
      </p>
    </div>
  {/if}
</nav>

<style lang="postcss">
  nav {
    @apply overflow-y-auto;
  }

  section {
    @apply flex flex-col gap-y-0.5 mb-3;
  }

  .section-header {
    @apply flex items-center gap-x-1.5 w-full;
    @apply px-2 py-1.5 cursor-pointer rounded-md;
    @apply text-fg-secondary;
  }

  .section-header:hover {
    @apply bg-surface-hover;
  }

  h3 {
    @apply text-[11px] uppercase tracking-wide font-semibold text-fg-muted;
  }

  ul {
    @apply flex flex-col gap-y-0.5 pl-1;
  }

  .row {
    @apply flex items-center gap-x-2 px-2 py-1.5 rounded-md;
    @apply text-sm text-fg-primary;
    @apply transition-colors;
  }

  .icon-wrap {
    @apply flex-none flex items-center justify-center;
    @apply size-[22px];
  }

  .row:hover {
    @apply bg-surface-hover;
  }

  .row.active {
    @apply bg-primary-100 text-primary-800 font-medium;
  }

  .empty-state {
    @apply flex flex-col gap-y-2 p-4 mt-4;
    @apply text-sm text-fg-muted;
  }

  .empty-state .hint {
    @apply text-xs;
  }

  .mono {
    @apply font-mono text-[11px] px-1 py-0.5 bg-surface-hover rounded;
  }
</style>
