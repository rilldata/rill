<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceListMagicAuthTokensInfinite,
    createAdminServiceRevokeMagicAuthToken,
    getAdminServiceListMagicAuthTokensQueryKey,
  } from "@rilldata/web-admin/client";
  import { useDashboards } from "@rilldata/web-admin/features/dashboards/listing/selectors";
  import NoPublicURLCTA from "@rilldata/web-admin/features/public-urls/NoPublicURLCTA.svelte";
  import PublicURLsTable from "@rilldata/web-admin/features/public-urls/PublicURLsTable.svelte";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";

  $: ({ instanceId } = $runtime);
  $: organization = $page.params.organization;
  $: project = $page.params.project;

  const PAGE_SIZE = 12;

  $: magicAuthTokensInfiniteQuery =
    createAdminServiceListMagicAuthTokensInfinite(
      organization,
      project,
      {
        pageSize: PAGE_SIZE,
      },
      {
        query: {
          getNextPageParam: (lastPage) => {
            if (lastPage.nextPageToken !== "") {
              return lastPage.nextPageToken;
            }
            return undefined;
          },
        },
      },
    );

  $: allRows =
    $magicAuthTokensInfiniteQuery.data?.pages.flatMap(
      (page) => page.tokens ?? [],
    ) ?? [];

  $: dashboards = useDashboards(instanceId);

  $: allRowsWithDashboardTitle = allRows.map((token) => {
    const dashboard = $dashboards.data?.find(
      (d) => d.meta.name.name === token.resourceName,
    );
    return {
      ...token,
      dashboardTitle: useValidDashboardTitle(dashboard),
    };
  });

  function useValidDashboardTitle(dashboard: V1Resource) {
    return dashboard?.explore?.spec?.displayName || dashboard?.meta.name.name;
  }

  // REVISIT when server-side sorting is implemented
  $: sortedAllRowsWithDashboardTitle = allRowsWithDashboardTitle.sort(
    (a, b) => new Date(b.createdOn).getTime() - new Date(a.createdOn).getTime(),
  );

  const queryClient = useQueryClient();
  const revokeMagicAuthToken = createAdminServiceRevokeMagicAuthToken();

  async function handleDelete(deletedTokenId: string) {
    try {
      await $revokeMagicAuthToken.mutateAsync({ tokenId: deletedTokenId });

      await queryClient.invalidateQueries({
        queryKey: getAdminServiceListMagicAuthTokensQueryKey(
          organization,
          project,
        ),
      });

      eventBus.emit("notification", { message: "Public URL deleted" });
    } catch {
      eventBus.emit("notification", {
        message: "Error deleting public URL",
        type: "error",
      });
    }
  }
</script>

<div class="flex flex-col w-full">
  <div class="flex md:flex-row flex-col gap-6">
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
