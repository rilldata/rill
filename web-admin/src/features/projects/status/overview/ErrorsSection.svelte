<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import {
    ResourceKind,
    SingletonProjectParserName,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { resourceIconMapping } from "@rilldata/web-common/features/entity-management/resource-icon-mapping";
  import {
    createRuntimeServiceGetResource,
    type V1Resource,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useResources } from "../selectors";
  import AlertCircleOutline from "@rilldata/web-common/components/icons/AlertCircleOutline.svelte";
  import { groupErrorsByKind } from "./overview-utils";

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

  // Total
  $: totalErrors = parseErrors.length + erroredResources.length;

  function handleSectionClick() {
    if (totalErrors > 0) {
      void goto(`${basePage}/resources?error=true`);
    }
  }
</script>

<div
  class="section"
  class:section-error={totalErrors > 0}
  class:section-clickable={totalErrors > 0}
  role={totalErrors > 0 ? "button" : undefined}
  tabindex={totalErrors > 0 ? 0 : undefined}
  on:click={handleSectionClick}
  on:keydown={(e) => {
    if (e.key === "Enter" || e.key === " ") {
      e.preventDefault();
      handleSectionClick();
    }
  }}
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
          href="{basePage}/resources?error=true"
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
</div>

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
