<script lang="ts">
  import { goto } from "$app/navigation";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import {
    resourceIconMapping,
    resourceLabelMapping,
    resourceShorthandMapping,
  } from "@rilldata/web-common/features/entity-management/resource-icon-mapping";
  import {
    coerceResourceKind,
    ResourceKind,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import ResourceTypeBadge from "@rilldata/web-common/features/entity-management/ResourceTypeBadge.svelte";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { ALLOWED_FOR_GRAPH } from "../navigation/seed-parser";

  export let resources: V1Resource[] = [];
  export let activeResourceId: string | null = null;

  type ResourceEntry = {
    name: string;
    kind: ResourceKind;
    displayKind: ResourceKind;
    status: "ok" | "pending" | "errored";
    resource: V1Resource;
  };

  type ResourceSection = {
    kind: ResourceKind;
    label: string;
    entries: ResourceEntry[];
  };

  const SECTION_ORDER: ResourceKind[] = [
    ResourceKind.Connector,
    ResourceKind.Source,
    ResourceKind.Model,
    ResourceKind.MetricsView,
    ResourceKind.Explore,
    ResourceKind.Canvas,
  ];

  const SECTION_LABELS: Partial<Record<ResourceKind, string>> = {
    [ResourceKind.Connector]: "OLAP Connector",
    [ResourceKind.Source]: "Source Models",
    [ResourceKind.Model]: "Models",
    [ResourceKind.MetricsView]: "Metric Views",
    [ResourceKind.Explore]: "Explore Dashboards",
    [ResourceKind.Canvas]: "Canvas Dashboards",
  };

  function getStatus(r: V1Resource): "ok" | "pending" | "errored" {
    if (r.meta?.reconcileError) return "errored";
    if (
      r.meta?.reconcileStatus &&
      r.meta.reconcileStatus !== "RECONCILE_STATUS_IDLE"
    )
      return "pending";
    return "ok";
  }

  $: sections = (function buildSections(): ResourceSection[] {
    const grouped = new Map<ResourceKind, ResourceEntry[]>();

    for (const r of resources) {
      const coerced = coerceResourceKind(r);
      // Allow connectors even if hidden; GraphContainer pre-filters to OLAP only
      if (r.meta?.hidden && coerced !== ResourceKind.Connector) continue;
      if (!coerced || !ALLOWED_FOR_GRAPH.has(coerced)) continue;

      const name = r.meta?.name?.name;
      if (!name) continue;

      const entry: ResourceEntry = {
        name,
        kind: r.meta?.name?.kind as ResourceKind,
        displayKind: coerced,
        status: getStatus(r),
        resource: r,
      };

      const existing = grouped.get(coerced) ?? [];
      existing.push(entry);
      grouped.set(coerced, existing);
    }

    const result: ResourceSection[] = [];
    for (const kind of SECTION_ORDER) {
      const entries = grouped.get(kind);
      if (!entries?.length) continue;
      entries.sort((a, b) => a.name.localeCompare(b.name));
      result.push({
        kind,
        label: SECTION_LABELS[kind] ?? resourceLabelMapping[kind] ?? "Unknown",
        entries,
      });
    }
    return result;
  })();

  $: totalCount = sections.reduce((sum, s) => sum + s.entries.length, 0);

  $: activeLabel = (function () {
    if (!activeResourceId) return `All resources (${totalCount})`;
    for (const section of sections) {
      for (const entry of section.entries) {
        const id = `${entry.kind}:${entry.name}`;
        if (id === activeResourceId) return entry.name;
      }
    }
    return `All resources (${totalCount})`;
  })();

  function handleSelect(entry: ResourceEntry) {
    const shortKind = resourceShorthandMapping[entry.kind];
    goto(`/graph?resource=${encodeURIComponent(`${shortKind}:${entry.name}`)}`);
  }

  function handleSelectAll() {
    goto("/graph");
  }
</script>

<div class="node-selector">
  <DropdownMenu.Root>
    <DropdownMenu.Trigger asChild let:builder>
      <button class="selector-trigger" use:builder.action {...builder}>
        <span class="trigger-label">{activeLabel}</span>
        <CaretDownIcon size="10px" />
      </button>
    </DropdownMenu.Trigger>
    <DropdownMenu.Content align="start" class="w-72">
      <DropdownMenu.Item on:click={handleSelectAll}>
        <span class="text-xs">All resources ({totalCount})</span>
      </DropdownMenu.Item>
      {#each sections as section}
        <DropdownMenu.Separator />
        <div class="section-header">
          <ResourceTypeBadge kind={section.kind} />
          <span class="text-[10px] text-fg-muted">{section.entries.length}</span
          >
        </div>
        {#each section.entries as entry}
          {@const isActive = activeResourceId === `${entry.kind}:${entry.name}`}
          <DropdownMenu.Item
            class="flex items-center gap-x-2 {isActive ? 'font-semibold' : ''}"
            on:click={() => handleSelect(entry)}
          >
            <svelte:component
              this={resourceIconMapping[entry.displayKind]}
              size="12px"
            />
            <span class="flex-1 truncate text-xs">{entry.name}</span>
            {#if entry.status === "errored"}
              <span class="status-dot errored"></span>
            {:else if entry.status === "pending"}
              <span class="status-dot pending"></span>
            {:else}
              <span class="status-dot ok"></span>
            {/if}
          </DropdownMenu.Item>
        {/each}
      {/each}
    </DropdownMenu.Content>
  </DropdownMenu.Root>
</div>

<style lang="postcss">
  .node-selector {
    @apply flex items-center;
  }

  .selector-trigger {
    @apply flex items-center gap-x-1.5 text-xs font-medium text-fg-secondary px-2 py-1 rounded transition-colors cursor-pointer;
  }

  .selector-trigger:hover {
    @apply text-fg-primary bg-surface-hover;
  }

  .trigger-label {
    @apply truncate max-w-[200px];
  }

  .section-header {
    @apply flex items-center justify-between px-2 py-1.5;
  }

  .status-dot {
    @apply flex-shrink-0 rounded-full;
    width: 6px;
    height: 6px;
  }

  .status-dot.ok {
    @apply bg-green-500;
  }

  .status-dot.pending {
    @apply bg-yellow-500;
  }

  .status-dot.errored {
    @apply bg-red-500;
  }
</style>
