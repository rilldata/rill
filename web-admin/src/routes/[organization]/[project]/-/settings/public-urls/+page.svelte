<script lang="ts">
  import PublicURLsTable from "@rilldata/web-admin/features/public-urls/PublicURLsTable.svelte";
  import { page } from "$app/stores";
  import {
    getAdminServiceListMagicAuthTokensQueryKey,
    createAdminServiceRevokeMagicAuthToken,
  } from "@rilldata/web-admin/client";
  import NoPublicURLCTA from "@rilldata/web-admin/features/public-urls/NoPublicURLCTA.svelte";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import { useDashboardsV2 } from "@rilldata/web-admin/features/dashboards/listing/selectors";
  import type { DashboardResource } from "@rilldata/web-admin/features/dashboards/listing/selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { createAdminServiceListMagicAuthTokensInfiniteQuery } from "@rilldata/web-admin/features/public-urls/create-infinite-query-public-urls";

  $: organization = $page.params.organization;
  $: project = $page.params.project;

  const PAGE_SIZE = 12;

  $: magicAuthTokensInfiniteQuery =
    createAdminServiceListMagicAuthTokensInfiniteQuery(organization, project, {
      pageSize: PAGE_SIZE,
    });

  function useValidDashboardTitle(dashboard: DashboardResource) {
    return (
      dashboard?.resource.metricsView.state?.validSpec?.title ||
      dashboard?.resource.meta.name.name
    );
  }

  $: allRows =
    $magicAuthTokensInfiniteQuery.data?.pages.flatMap(
      (page) => page.tokens ?? [],
    ) ?? [];

  $: dashboards = useDashboardsV2($runtime.instanceId);

  $: allRowsWithDashboardTitle = allRows.map((token) => {
    const dashboard = $dashboards.data?.find(
      (d) => d.resource.meta.name.name === token.metricsView,
    );
    return {
      ...token,
      dashboardTitle: useValidDashboardTitle(dashboard),
    };
  });

  // REVISIT when server-side sorting is implemented
  $: sortedAllRowsWithDashboardTitle = allRowsWithDashboardTitle.sort(
    (a, b) => new Date(b.createdOn).getTime() - new Date(a.createdOn).getTime(),
  );

  const queryClient = useQueryClient();
  const revokeMagicAuthToken = createAdminServiceRevokeMagicAuthToken();

  async function handleDelete(deletedTokenId: string) {
    try {
      await $revokeMagicAuthToken.mutateAsync({ tokenId: deletedTokenId });

      // Invalidate and refetch the query
      await queryClient.invalidateQueries(
        getAdminServiceListMagicAuthTokensQueryKey(organization, project),
      );

      eventBus.emit("notification", { message: "Public URL deleted" });
    } catch (error) {
      eventBus.emit("notification", {
        message: "Error deleting public URL",
        type: "error",
      });
    }
  }
</script>

<div class="flex flex-col w-full">
  <div class="flex md:flex-row flex-col gap-6">
    <div class="w-full">
      {#if $magicAuthTokensInfiniteQuery.isLoading}
        <DelayedSpinner
          isLoading={$magicAuthTokensInfiniteQuery.isLoading}
          size="1rem"
        />
      {:else if $magicAuthTokensInfiniteQuery.isError}
        <div class="text-red-500">
          Error loading public URLs: {$magicAuthTokensInfiniteQuery.error}
        </div>
      {:else if $magicAuthTokensInfiniteQuery.isSuccess}
        {#if $magicAuthTokensInfiniteQuery.data.pages[0].tokens.length === 0}
          <NoPublicURLCTA />
        {:else}
          <PublicURLsTable
            data={sortedAllRowsWithDashboardTitle}
            query={$magicAuthTokensInfiniteQuery}
            onDelete={handleDelete}
          />
        {/if}
      {/if}
    </div>
  </div>
</div>
