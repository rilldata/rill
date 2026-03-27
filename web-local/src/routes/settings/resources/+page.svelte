<script lang="ts">
  import { page } from "$app/stores";
  import { onMount } from "svelte";
  import ResourcesFilterableTable from "@rilldata/web-common/features/resources/ResourcesFilterableTable.svelte";
  import ParseErrorsSection from "../../status/ParseErrorsSection.svelte";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import {
    createRuntimeServiceCreateTriggerMutation,
    createRuntimeServiceListResources,
    getRuntimeServiceListResourcesQueryKey,
    V1ReconcileStatus,
  } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { useQueryClient } from "@tanstack/svelte-query";
  import {
    createUrlFilterSync,
    parseArrayParam,
    parseStringParam,
  } from "@rilldata/web-common/lib/url-filter-sync";

  const runtimeClient = useRuntimeClient();
  const queryClient = useQueryClient();
  const createTrigger =
    createRuntimeServiceCreateTriggerMutation(runtimeClient);

  const filterSync = createUrlFilterSync([
    { key: "kind", type: "array" },
    { key: "status", type: "array" },
    { key: "q", type: "string" },
  ]);
  filterSync.init($page.url);

  let searchText = parseStringParam($page.url.searchParams.get("q"));
  let selectedTypes = parseArrayParam($page.url.searchParams.get("kind"));
  let selectedStatuses = parseArrayParam($page.url.searchParams.get("status"));
  let mounted = false;

  $: if (mounted && filterSync.hasExternalNavigation($page.url)) {
    filterSync.markSynced($page.url);
    selectedTypes = parseArrayParam($page.url.searchParams.get("kind"));
    selectedStatuses = parseArrayParam($page.url.searchParams.get("status"));
    searchText = parseStringParam($page.url.searchParams.get("q"));
  }

  $: if (mounted) {
    filterSync.syncToUrl({
      kind: selectedTypes,
      status: selectedStatuses,
      q: searchText,
    });
  }

  onMount(() => {
    mounted = true;
  });

  $: resourcesQuery = createRuntimeServiceListResources(
    runtimeClient,
    {},
    { query: { refetchInterval: 5000 } },
  );

  $: resources = $resourcesQuery.data?.resources ?? [];

  $: hasReconcilingSourcesOrModels = resources.some(
    (r) =>
      (r.meta?.name?.kind === ResourceKind.Source ||
        r.meta?.name?.kind === ResourceKind.Model) &&
      (r.meta?.reconcileStatus === V1ReconcileStatus.RECONCILE_STATUS_PENDING ||
        r.meta?.reconcileStatus === V1ReconcileStatus.RECONCILE_STATUS_RUNNING),
  );

  function refreshAllSourcesAndModels() {
    void $createTrigger.mutateAsync({ all: true }).then(() => {
      void queryClient.invalidateQueries({
        queryKey: getRuntimeServiceListResourcesQueryKey(
          runtimeClient.instanceId,
          undefined,
        ),
      });
    });
  }
</script>

<svelte:head>
  <title>Rill Developer | Resources</title>
</svelte:head>

<ResourcesFilterableTable
  {resources}
  containerHeight={550}
  isLoading={$resourcesQuery.isLoading}
  isError={$resourcesQuery.isError}
  errorMessage={$resourcesQuery.error?.message ?? ""}
  isRefreshDisabled={hasReconcilingSourcesOrModels}
  onRefreshAll={refreshAllSourcesAndModels}
  onRefetch={() => $resourcesQuery.refetch()}
  bind:selectedStatuses
  bind:selectedTypes
  bind:searchText
/>

<ParseErrorsSection />
