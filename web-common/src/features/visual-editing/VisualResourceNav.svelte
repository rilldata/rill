<script lang="ts">
  import { page } from "$app/stores";
  import CanvasIcon from "@rilldata/web-common/components/icons/CanvasIcon.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import ExploreIcon from "@rilldata/web-common/components/icons/ExploreIcon.svelte";
  import MetricsViewIcon from "@rilldata/web-common/components/icons/MetricsViewIcon.svelte";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { createRuntimeServiceListResources } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";

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

  function hrefFor(filePath: string | undefined): string | undefined {
    if (!filePath) return undefined;
    return `/files${filePath.startsWith("/") ? "" : "/"}${filePath}`;
  }

  type Item = {
    name: string;
    filePath: string | undefined;
    href: string | undefined;
  };

  function itemsFrom(
    resources: Array<{
      meta?: { name?: { name?: string }; filePaths?: string[] };
    }>,
  ): Item[] {
    return resources
      .map((r) => {
        const filePath = r.meta?.filePaths?.[0];
        return {
          name: r.meta?.name?.name ?? "",
          filePath,
          href: hrefFor(filePath),
        };
      })
      .filter((i) => i.name && i.href)
      .sort((a, b) =>
        a.name.localeCompare(b.name, undefined, { sensitivity: "base" }),
      );
  }

  $: exploreItems = itemsFrom(explores);
  $: canvasItems = itemsFrom(canvases);
  $: metricsItems = itemsFrom(metrics);

  let dashboardsOpen = true;
  let canvasesOpen = true;
  let metricsOpen = true;
</script>

<nav class="flex flex-col gap-y-1 p-2 pb-6 w-full">
  {#if exploreItems.length}
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
          {#each exploreItems as item (item.href)}
            <li>
              <a
                href={item.href}
                class="row"
                class:active={currentFile === item.filePath}
              >
                <ExploreIcon size="22px" />
                <span class="truncate">{item.name}</span>
              </a>
            </li>
          {/each}
        </ul>
      {/if}
    </section>
  {/if}

  {#if canvasItems.length}
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
          {#each canvasItems as item (item.href)}
            <li>
              <a
                href={item.href}
                class="row"
                class:active={currentFile === item.filePath}
              >
                <CanvasIcon size="22px" />
                <span class="truncate">{item.name}</span>
              </a>
            </li>
          {/each}
        </ul>
      {/if}
    </section>
  {/if}

  {#if metricsItems.length}
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
          {#each metricsItems as item (item.href)}
            <li>
              <a
                href={item.href}
                class="row"
                class:active={currentFile === item.filePath}
              >
                <MetricsViewIcon size="22px" />
                <span class="truncate">{item.name}</span>
              </a>
            </li>
          {/each}
        </ul>
      {/if}
    </section>
  {/if}

  {#if !exploreItems.length && !canvasItems.length && !metricsItems.length}
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
