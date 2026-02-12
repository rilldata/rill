<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import {
    ResourceKind,
    SingletonProjectParserName,
    prettyResourceKind,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { resourceIconMapping } from "@rilldata/web-common/features/entity-management/resource-icon-mapping";
  import {
    createRuntimeServiceGetResource,
    type V1Resource,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useResources } from "../selectors";
  import AlertCircleOutline from "@rilldata/web-common/components/icons/AlertCircleOutline.svelte";

  $: ({ instanceId } = $runtime);
  $: basePage = `/${$page.params.organization}/${$page.params.project}/-/status`;

  // Parse errors
  $: projectParserQuery = createRuntimeServiceGetResource(
    instanceId,
    {
      "name.kind": ResourceKind.ProjectParser,
      "name.name": SingletonProjectParserName,
    },
    { query: { refetchOnMount: true, refetchOnWindowFocus: true } },
  );
  $: parseErrors =
    $projectParserQuery.data?.resource?.projectParser?.state?.parseErrors ?? [];

  // Resource errors grouped by kind
  $: resourcesQuery = useResources(instanceId);
  $: allResources = ($resourcesQuery.data?.resources ?? []) as V1Resource[];
  $: erroredResources = allResources.filter((r) => !!r.meta?.reconcileError);

  $: errorsByKind = groupErrorsByKind(erroredResources);

  function groupErrorsByKind(
    resources: V1Resource[],
  ): { kind: string; label: string; count: number }[] {
    const counts = new Map<string, number>();
    for (const r of resources) {
      const kind = r.meta?.name?.kind;
      if (kind) counts.set(kind, (counts.get(kind) ?? 0) + 1);
    }
    return Array.from(counts.entries())
      .map(([kind, count]) => ({
        kind,
        label: prettyResourceKind(kind),
        count,
      }))
      .sort((a, b) => b.count - a.count);
  }

  // Total
  $: totalErrors = parseErrors.length + erroredResources.length;

  function handleSectionClick() {
    if (totalErrors > 0) {
      void goto(`${basePage}/resources?error=true`);
    }
  }
</script>

<!-- svelte-ignore a11y-click-events-have-key-events -->
<!-- svelte-ignore a11y-no-static-element-interactions -->
<section
  class="section"
  class:section-error={totalErrors > 0}
  class:section-clickable={totalErrors > 0}
  on:click={handleSectionClick}
>
  <div class="section-header">
    <h3 class="section-title flex items-center gap-2">
      Errors
      {#if totalErrors > 0}
        <span class="error-badge">{totalErrors}</span>
      {/if}
    </h3>
  </div>

  {#if totalErrors === 0}
    <p class="text-sm text-fg-secondary">No errors detected.</p>
  {:else}
    <div class="error-chips">
      {#if parseErrors.length > 0}
        <a
          href="{basePage}/project-logs"
          class="error-chip"
          on:click|stopPropagation
        >
          <AlertCircleOutline size="12px" />
          <span class="font-medium">{parseErrors.length}</span>
          <span>Parse error{parseErrors.length !== 1 ? "s" : ""}</span>
        </a>
      {/if}

      {#each errorsByKind as { kind, label, count }}
        <a
          href="{basePage}/resources?error=true&kind={kind}"
          class="error-chip"
          on:click|stopPropagation
        >
          {#if resourceIconMapping[kind]}
            <svelte:component this={resourceIconMapping[kind]} size="12px" />
          {/if}
          <span class="font-medium">{count}</span>
          <span>{label}{count !== 1 ? "s" : ""}</span>
        </a>
      {/each}
    </div>
  {/if}
</section>

<style lang="postcss">
  .section {
    @apply block border border-border rounded-lg p-5;
  }
  .section-clickable {
    @apply cursor-pointer;
  }
  .section-error {
    @apply border-red-500;
  }
  .section-clickable:hover {
    @apply border-red-600;
  }
  .section-header {
    @apply flex items-center justify-between mb-4;
  }
  .section-title {
    @apply text-sm font-semibold text-fg-primary uppercase tracking-wide;
  }
  .error-badge {
    @apply text-xs font-semibold text-white bg-red-500 rounded-full px-1.5 py-0.5 min-w-[20px] text-center;
  }
  .error-chips {
    @apply flex flex-wrap gap-2;
  }
  .error-chip {
    @apply flex items-center gap-1.5 text-xs px-2.5 py-1.5 rounded-md;
    @apply border border-red-300 bg-red-50 text-red-700 no-underline;
  }
  .error-chip:hover {
    @apply border-red-500 bg-red-100;
  }
</style>
