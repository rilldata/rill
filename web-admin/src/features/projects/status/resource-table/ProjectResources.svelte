<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import ResourcesFilterableTable from "@rilldata/web-common/features/resources/ResourcesFilterableTable.svelte";
  import ParseErrorsSection from "@rilldata/web-common/features/resources/ParseErrorsSection.svelte";
  import {
    ResourceKind,
    SingletonProjectParserName,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import {
    createRuntimeServiceCreateTrigger,
    createRuntimeServiceGetResource,
    getRuntimeServiceListResourcesQueryKey,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { useResources } from "../selectors";
  import { isResourceReconciling } from "@rilldata/web-admin/lib/refetch-interval-store";
  import { onMount } from "svelte";

  const queryClient = useQueryClient();
  const createTrigger = createRuntimeServiceCreateTrigger();

  // Reactively track URL search params (updates on back/forward navigation)
  $: kindParam = $page.url.searchParams.get("kind");
  $: statusParam = $page.url.searchParams.get("status");
  $: qParam = $page.url.searchParams.get("q");

  let searchText = $page.url.searchParams.get("q") ?? "";
  let selectedTypes: string[] = (() => {
    const k = $page.url.searchParams.get("kind");
    return k ? k.split(",").filter(Boolean) : [];
  })();
  let selectedStatuses: string[] = (() => {
    const s = $page.url.searchParams.get("status");
    return s ? s.split(",").filter(Boolean) : [];
  })();
  let mounted = false;
  let lastSyncedSearch = $page.url.search;

  // Sync URL → local state on external navigation (back/forward)
  $: if (mounted && $page.url.search !== lastSyncedSearch) {
    lastSyncedSearch = $page.url.search;
    selectedTypes = kindParam ? kindParam.split(",").filter(Boolean) : [];
    selectedStatuses = statusParam
      ? statusParam.split(",").filter(Boolean)
      : [];
    searchText = qParam ?? "";
  }

  // Sync filter state to URL params
  $: if (mounted) {
    syncFiltersToUrl(selectedTypes, selectedStatuses, searchText);
  }

  onMount(() => {
    mounted = true;
  });

  function syncFiltersToUrl(
    types: string[],
    statuses: string[],
    search: string,
  ) {
    const url = new URL($page.url);
    if (types.length > 0) {
      url.searchParams.set("kind", types.join(","));
    } else {
      url.searchParams.delete("kind");
    }
    if (statuses.length > 0) {
      url.searchParams.set("status", statuses.join(","));
    } else {
      url.searchParams.delete("status");
    }
    if (search) {
      url.searchParams.set("q", search);
    } else {
      url.searchParams.delete("q");
    }
    lastSyncedSearch = url.search;
    void goto(url.pathname + url.search, {
      replaceState: true,
      noScroll: true,
      keepFocus: true,
    });
  }

  $: ({ instanceId } = $runtime);

  $: resources = useResources(instanceId);

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

  $: hasReconcilingResources = $resources.data?.resources?.some(
    isResourceReconciling,
  );

  function refreshAllSourcesAndModels() {
    void $createTrigger
      .mutateAsync({
        instanceId,
        data: { all: true },
      })
      .then(() => {
        void queryClient.invalidateQueries({
          queryKey: getRuntimeServiceListResourcesQueryKey(
            instanceId,
            undefined,
          ),
        });
      });
  }
</script>

<ResourcesFilterableTable
  resources={$resources.data?.resources ?? []}
  isLoading={$resources.isLoading}
  isError={$resources.isError}
  errorMessage={$resources.error?.message ?? ""}
  isRefreshDisabled={hasReconcilingResources ?? false}
  onRefreshAll={refreshAllSourcesAndModels}
  onRefetch={() => $resources.refetch()}
  bind:selectedStatuses
  bind:selectedTypes
  bind:searchText
  containerHeight={550}
  emptyText="No resources match the current filters"
>
  <svelte:fragment slot="after-table">
    <div class="pt-4 mt-2">
      <ParseErrorsSection {parseErrors} />
    </div>
  </svelte:fragment>
</ResourcesFilterableTable>
