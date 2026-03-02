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
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { useResources } from "../selectors";
  import AlertCircleOutline from "@rilldata/web-common/components/icons/AlertCircleOutline.svelte";
  import { groupErrorsByKind, pluralizeKind } from "./overview-utils";

  const runtimeClient = useRuntimeClient();
  $: basePage = `/${$page.params.organization}/${$page.params.project}/-/status`;

  // Parse errors
  $: projectParserQuery = createRuntimeServiceGetResource(
    runtimeClient,
    {
      name: {
        kind: ResourceKind.ProjectParser,
        name: SingletonProjectParserName,
      },
    },
    { query: { refetchOnMount: true, refetchOnWindowFocus: true } },
  );
  $: parseErrors =
    $projectParserQuery.data?.resource?.projectParser?.state?.parseErrors ?? [];

  // Resource errors grouped by kind
  $: resourcesQuery = useResources(runtimeClient);
  $: allResources = ($resourcesQuery.data?.resources ?? []) as V1Resource[];
  $: erroredResources = allResources.filter((r) => !!r.meta?.reconcileError);

  $: errorsByKind = groupErrorsByKind(erroredResources);

  // Total
  $: totalErrors = parseErrors.length + erroredResources.length;

  function handleSectionClick(e: MouseEvent | KeyboardEvent) {
    // Don't navigate if the click was on a chip link
    if ((e.target as HTMLElement).closest(".error-chip")) return;
    if (totalErrors > 0) {
      void goto(`${basePage}/resources?status=error`);
    }
  }
</script>

{#if totalErrors > 0}
  <div
    class="section section-error section-clickable"
    role="button"
    tabindex="0"
    on:click={handleSectionClick}
    on:keydown={(e) => {
      if (e.key === "Enter" || e.key === " ") {
        e.preventDefault();
        handleSectionClick(e);
      }
    }}
  >
    <div class="section-header">
      <h3 class="section-title flex items-center gap-2">
        Errors
        <span class="error-badge">{totalErrors}</span>
      </h3>
    </div>

    <div class="error-chips">
      {#if parseErrors.length > 0}
        <a href="{basePage}/resources?status=error" class="error-chip">
          <AlertCircleOutline size="12px" />
          <span class="font-medium">{parseErrors.length}</span>
          <span>Parse error{parseErrors.length !== 1 ? "s" : ""}</span>
        </a>
      {/if}

      {#each errorsByKind as { kind, label, count } (kind)}
        <a
          href="{basePage}/resources?status=error&kind={kind}"
          class="error-chip"
        >
          {#if resourceIconMapping[kind]}
            <svelte:component this={resourceIconMapping[kind]} size="12px" />
          {/if}
          <span class="font-medium">{count}</span>
          <span>{pluralizeKind(label, count)}</span>
        </a>
      {/each}
    </div>
  </div>
{:else}
  <div class="section">
    <div class="section-header">
      <h3 class="section-title flex items-center gap-2">Errors</h3>
    </div>

    {#if $projectParserQuery.isError || $resourcesQuery.isError}
      <p class="text-sm text-fg-secondary">Unable to check for errors.</p>
    {:else if $projectParserQuery.isLoading || $resourcesQuery.isLoading}
      <p class="text-sm text-fg-secondary">Checking for errors...</p>
    {:else}
      <p class="text-sm text-fg-secondary">No errors detected.</p>
    {/if}
  </div>
{/if}

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
