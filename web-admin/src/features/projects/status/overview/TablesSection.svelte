<script lang="ts">
  import { page } from "$app/stores";
  import { createRuntimeServiceGetInstance } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { useInfiniteTablesList } from "../selectors";
  import {
    filterTemporaryTables,
    isLikelyView,
  } from "@rilldata/web-common/features/projects/status/tables/utils";
  import { writable } from "svelte/store";
  import OverviewCard from "@rilldata/web-common/features/projects/status/overview/OverviewCard.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  const runtimeClient = useRuntimeClient();
  $: basePage = `/${$page.params.organization}/${$page.params.project}/-/status`;

  // Get instance info for OLAP connector
  $: instanceQuery = createRuntimeServiceGetInstance(runtimeClient, {
    sensitive: true,
  });
  $: instance = $instanceQuery.data?.instance;

  // Get tables list (first page only; show "+" suffix when more pages exist)
  $: connectorName = instance?.olapConnector ?? "";
  const tablesParams = writable({
    client: runtimeClient,
    connector: "",
    searchPattern: undefined as string | undefined,
  });
  $: tablesParams.set({
    client: runtimeClient,
    connector: connectorName,
    searchPattern: undefined,
  });
  const tablesList = useInfiniteTablesList(tablesParams);

  $: filteredTables = filterTemporaryTables($tablesList.data?.tables);
  $: hasMore = $tablesList.hasNextPage;

  // Count tables vs views using size heuristic (same logic as tables page).
  // isLikelyView returns undefined when indeterminate; treat as table for counts.
  $: viewCount = filteredTables.filter(
    (t) => isLikelyView(undefined, t.physicalSizeBytes) === true,
  ).length;
  $: tableCount = filteredTables.length - viewCount;

  // Show loading when prerequisites aren't ready OR doing an initial fetch (no cache).
  // `isLoading && isFetching` is TanStack Query's "hard loading" pattern:
  // it's false when cached data exists (even if refetching in background).
  $: isLoading =
    $instanceQuery.isLoading ||
    !connectorName ||
    ($tablesList.isLoading && $tablesList.isFetching);
</script>

<OverviewCard title={m.status_nav_tables()} viewAllHref="{basePage}/tables">
  {#if isLoading}
    <p class="text-sm text-fg-secondary">{m.status_loading_tables()}</p>
  {:else if filteredTables.length > 0}
    <div class="chips">
      <a href="{basePage}/tables?type=table" class="chip">
        <span class="font-medium"
          >{tableCount}{hasMore && tableCount > 0 ? "+" : ""}</span
        >
        <span class="text-fg-secondary"
          >{m.status_table_label({ count: tableCount })}</span
        >
      </a>
      <a href="{basePage}/tables?type=view" class="chip">
        <span class="font-medium"
          >{viewCount}{hasMore && viewCount > 0 ? "+" : ""}</span
        >
        <span class="text-fg-secondary"
          >{m.status_view_label({ count: viewCount })}</span
        >
      </a>
    </div>
  {:else}
    <p class="text-sm text-fg-secondary">{m.status_no_tables()}</p>
  {/if}
</OverviewCard>

<style lang="postcss">
  .chips {
    @apply flex flex-wrap gap-2;
  }
  .chip {
    @apply flex items-center gap-1.5 text-xs px-2.5 py-1.5 rounded-md border border-border bg-surface-subtle no-underline text-inherit;
  }
  .chip:hover {
    @apply border-primary-500 text-primary-600;
  }
</style>
