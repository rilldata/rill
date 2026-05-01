<script lang="ts">
  import { page } from "$app/stores";
  import {
    ResourceKind,
    SingletonProjectParserName,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import ParseErrorsSection from "@rilldata/web-common/features/resources/ParseErrorsSection.svelte";
  import ResourcesFilterableTable from "@rilldata/web-common/features/resources/ResourcesFilterableTable.svelte";
  import {
    createRuntimeServiceCreateTriggerMutation,
    createRuntimeServiceGetResource,
    createRuntimeServiceListResources,
    getRuntimeServiceListResourcesQueryKey,
    V1ReconcileStatus,
  } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import {
    createUrlFilterSync,
    parseArrayParam,
    parseStringParam,
  } from "@rilldata/web-common/lib/url-filter-sync";
  import { useQueryClient } from "@tanstack/svelte-query";

  const runtimeClient = useRuntimeClient();
  const queryClient = useQueryClient();
  const createTrigger =
    createRuntimeServiceCreateTriggerMutation(runtimeClient);

  const filterSync = createUrlFilterSync([
    { key: "kind", type: "array" },
    { key: "status", type: "array" },
    { key: "q", type: "string" },
  ]);
  $: kindFilter = parseArrayParam($page.url.searchParams.get("kind"));
  $: statusFilter = parseArrayParam($page.url.searchParams.get("status"));
  $: search = parseStringParam($page.url.searchParams.get("q"));

  $: resourcesQuery = createRuntimeServiceListResources(runtimeClient, {});
  $: resources = $resourcesQuery.data?.resources ?? [];

  $: projectParserQuery = createRuntimeServiceGetResource(runtimeClient, {
    name: {
      kind: ResourceKind.ProjectParser,
      name: SingletonProjectParserName,
    },
  });
  $: parseErrors =
    $projectParserQuery.data?.resource?.projectParser?.state?.parseErrors ?? [];
</script>

{#if parseErrors.length > 0}
  <ParseErrorsSection {parseErrors} />
{/if}

<ResourcesFilterableTable
  {resources}
  {kindFilter}
  {statusFilter}
  {search}
  reconcileInFlight={resources.some(
    (r) =>
      r.meta?.reconcileStatus === V1ReconcileStatus.RECONCILE_STATUS_RUNNING,
  )}
  onFilterChange={(next) => filterSync.update(next)}
  onTrigger={async (names) => {
    await $createTrigger.mutateAsync({
      instanceId: runtimeClient.instanceId,
      data: { resources: names },
    });
    void queryClient.invalidateQueries({
      queryKey: getRuntimeServiceListResourcesQueryKey(
        runtimeClient.instanceId,
        {},
      ),
    });
  }}
/>
