<script lang="ts">
  import { page } from "$app/stores";
  import AlertCircleOutline from "@rilldata/web-common/components/icons/AlertCircleOutline.svelte";
  import ChevronRight from "@rilldata/web-common/components/icons/ChevronRight.svelte";
  import {
    ResourceKind,
    SingletonProjectParserName,
    prettyResourceKind,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import {
    createRuntimeServiceGetResource,
    type V1Resource,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useResources } from "../selectors";

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
  // Resource errors
  $: resourcesQuery = useResources(instanceId);
  $: allResources = ($resourcesQuery.data?.resources ?? []) as V1Resource[];
  $: erroredResources = allResources.filter((r) => !!r.meta?.reconcileError);

  // Total
  $: totalErrors = parseErrors.length + erroredResources.length;
</script>

<section class="section" class:section-error={totalErrors > 0}>
  <div class="section-header">
    <h3 class="section-title flex items-center gap-2">
      Errors
      {#if totalErrors > 0}
        <span class="error-badge">{totalErrors}</span>
      {/if}
    </h3>
    {#if totalErrors > 0}
      <a href="{basePage}/resources?error=true" class="section-action">
        View all errors
        <ChevronRight size="14px" />
      </a>
    {/if}
  </div>

  {#if totalErrors === 0}
    <p class="text-sm text-fg-secondary">No errors detected.</p>
  {:else}
    <div class="error-list">
      {#if parseErrors.length > 0}
        <div class="error-item">
          <AlertCircleOutline size="16px" color="#ef4444" />
          <span class="font-medium"
            >{parseErrors.length} parse error{parseErrors.length !== 1
              ? "s"
              : ""}</span
          >
        </div>
      {/if}

      {#if erroredResources.length > 0}
        <div class="error-item">
          <AlertCircleOutline size="16px" color="#ef4444" />
          <span class="font-medium"
            >{erroredResources.length} resource error{erroredResources.length !==
            1
              ? "s"
              : ""}</span
          >
        </div>
      {/if}
    </div>
  {/if}
</section>

<style lang="postcss">
  .section {
    @apply border border-border rounded-lg p-5;
  }
  .section-error {
    @apply border-red-500;
  }
  .section-header {
    @apply flex items-center justify-between mb-4;
  }
  .section-title {
    @apply text-sm font-semibold text-fg-primary uppercase tracking-wide;
  }
  .section-action {
    @apply text-sm text-primary-500 flex items-center gap-1;
  }
  .section-action:hover {
    @apply text-primary-600;
  }
  .error-badge {
    @apply text-xs font-semibold text-white bg-red-500 rounded-full px-1.5 py-0.5 min-w-[20px] text-center;
  }
  .error-list {
    @apply flex flex-col gap-1;
  }
  .error-item {
    @apply flex items-center gap-3 px-3 py-2.5 rounded-md text-sm;
  }
  .error-item:hover {
    @apply bg-surface-subtle;
  }
</style>
