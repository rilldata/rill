<script lang="ts">
  import ResourceGraphContainer from "@rilldata/web-common/features/resource-graph/ResourceGraphContainer.svelte";
  import WorkspaceContainer from "@rilldata/web-common/layout/workspace/WorkspaceContainer.svelte";
  import StatsHeader from "@rilldata/web-common/features/resource-graph/StatsHeader.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import { page } from "$app/stores";
  import { goto } from "$app/navigation";

  // Collect `seed` params from URL (supports multiple)
  $: seeds = Array.from($page.url.searchParams.getAll("seed"));

  // Derive a simple mode from the current URL seeds
  const KIND_TOKENS = new Set(["metrics", "sources", "models", "dashboards"]);
  $: seedMode = (() => {
    if (!seeds.length) return "metrics"; // default view
    if (seeds.length === 1) {
      const token = (seeds[0] || "").toLowerCase();
      if (KIND_TOKENS.has(token)) return token;
    }
    return "metrics"; // fallback to metrics selection for custom URLs
  })();

  const seedOptions = [
    { value: "metrics", label: "Metrics" },
    { value: "sources", label: "Sources" },
    { value: "models", label: "Models" },
    { value: "dashboards", label: "Dashboards" },
  ];

  function setSeedMode(mode: string) {
    const url = new URL($page.url);
    url.searchParams.delete("seed");
    url.searchParams.append("seed", mode);
    const qs = url.searchParams.toString();
    const newPath = qs ? `${url.pathname}?${qs}` : url.pathname;
    goto(newPath);
  }
</script>

<svelte:head>
  <title>Rill Developer | Project graph</title>
</svelte:head>

<WorkspaceContainer inspector={false}>
  <div slot="header" class="header">
    <div class="header-title">
      <div class="header-left">
        <h1>Project graph</h1>
      </div>
      <div class="header-right">
        <StatsHeader {seeds} />
        <span class="seed-label">Source component:</span>
        <Select
          id="graph-seed-mode"
          ariaLabel="Graph seed mode"
          size="md"
          minWidth={220}
          dropdownWidth="min-w-[260px]"
          onChange={setSeedMode}
          options={seedOptions}
          bind:value={seedMode}
        />
      </div>
    </div>
    <p>Visualize dependencies between sources, models, dashboards, and more.</p>
  </div>

  <div slot="body" class="graph-wrapper">
    <ResourceGraphContainer seeds={seeds} />
  </div>
</WorkspaceContainer>

<style lang="postcss">
  .header {
    @apply px-4 pt-3 pb-2;
  }

  .header h1 {
    @apply text-lg font-semibold text-foreground;
  }

  .header-title {
    @apply flex items-center justify-between;
  }
  .header-right {
    @apply flex items-center gap-x-2 ml-3;
  }
  .seed-label {
    @apply text-xs text-gray-600 ml-2;
  }

  .header p {
    @apply text-sm text-gray-500 mt-1;
  }

  .graph-wrapper {
    @apply h-full w-full;
  }
</style>
