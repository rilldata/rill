<script lang="ts">
  import { goto } from "$app/navigation";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { resourceShorthandMapping } from "@rilldata/web-common/features/entity-management/resource-icon-mapping";

  export let connectors = 0;
  export let sources = 0;
  export let models = 0;
  export let metrics = 0;
  export let dashboards = 0;
  export let activeToken:
    | "connectors"
    | "sources"
    | "metrics"
    | "models"
    | "dashboards"
    | null = null;

  $: total = connectors + sources + models + metrics + dashboards;

  const kindConfig = [
    {
      token: "connectors" as const,
      label: "Connectors",
      kind: ResourceKind.Connector,
    },
    {
      token: "sources" as const,
      label: "Sources",
      kind: ResourceKind.Source,
    },
    {
      token: "models" as const,
      label: "Models",
      kind: ResourceKind.Model,
    },
    {
      token: "metrics" as const,
      label: "Metrics",
      kind: ResourceKind.MetricsView,
    },
    {
      token: "dashboards" as const,
      label: "Dashboards",
      kind: ResourceKind.Explore,
    },
  ];

  function getCount(
    token: "connectors" | "sources" | "models" | "metrics" | "dashboards",
  ): number {
    switch (token) {
      case "connectors":
        return connectors;
      case "sources":
        return sources;
      case "models":
        return models;
      case "metrics":
        return metrics;
      case "dashboards":
        return dashboards;
    }
  }

  function handleChange(e: Event) {
    const val = (e.currentTarget as HTMLSelectElement).value;
    if (val === "__all__") {
      goto("/graph");
    } else {
      goto(`/graph?kind=${val}`);
    }
  }
</script>

<div class="kind-selector">
  <span class="kind-label">Node:</span>
  <select
    class="kind-dropdown"
    value={activeToken ?? "__all__"}
    on:change={handleChange}
  >
    <option value="__all__">All resources ({total})</option>
    {#each kindConfig as config}
      {@const count = getCount(config.token)}
      <option value={config.token} disabled={count === 0}>
        {config.label} ({count})
      </option>
    {/each}
  </select>
</div>

<style lang="postcss">
  .kind-selector {
    @apply flex items-center gap-1.5 mb-2;
  }

  .kind-label {
    @apply text-xs font-medium text-fg-secondary;
  }

  .kind-dropdown {
    @apply h-7 px-2.5 pr-7 text-xs rounded border bg-surface-background text-fg-primary;
    @apply focus:outline-none focus:ring-1 focus:ring-primary-500 focus:border-primary-500;
    @apply cursor-pointer appearance-none;
    background-image: url("data:image/svg+xml,%3csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 20 20'%3e%3cpath stroke='%236b7280' stroke-linecap='round' stroke-linejoin='round' stroke-width='1.5' d='M6 8l4 4 4-4'/%3e%3c/svg%3e");
    background-position: right 0.25rem center;
    background-repeat: no-repeat;
    background-size: 1.25rem 1.25rem;
  }
</style>
