<script lang="ts">
  import { goto } from "$app/navigation";

  import type { KindToken } from "../navigation/seed-parser";

  export let connector = 0;
  export let sources = 0;
  export let models = 0;
  export let metrics = 0;
  export let dashboards = 0;
  export let activeToken: KindToken | null = null;

  $: total = connector + sources + models + metrics + dashboards;

  const kindConfig: { token: KindToken; label: string }[] = [
    { token: "connector", label: "OLAP Connector" },
    { token: "sources", label: "Source Models" },
    { token: "models", label: "Models" },
    { token: "metrics", label: "Metric Views" },
    { token: "dashboards", label: "Dashboards" },
  ];

  function getCount(token: KindToken): number {
    switch (token) {
      case "connector":
        return connector;
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
    @apply cursor-pointer appearance-none;
    background-image: url("data:image/svg+xml,%3csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 20 20'%3e%3cpath stroke='%236b7280' stroke-linecap='round' stroke-linejoin='round' stroke-width='1.5' d='M6 8l4 4 4-4'/%3e%3c/svg%3e");
    background-position: right 0.25rem center;
    background-repeat: no-repeat;
    background-size: 1.25rem 1.25rem;
  }

  .kind-dropdown:focus {
    @apply outline-none ring-1 ring-primary-500 border-primary-500;
  }
</style>
