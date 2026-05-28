<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceListMagicAuthTokensInfinite,
    createAdminServiceRevokeMagicAuthToken,
    getAdminServiceListMagicAuthTokensQueryKey,
  } from "@rilldata/web-admin/client";
  import { useDashboards } from "@rilldata/web-admin/features/dashboards/listing/selectors";
  import PublicURLsResourceTable from "@rilldata/web-admin/features/public-urls/PublicURLsResourceTable.svelte";
  import RadixLarge from "@rilldata/web-common/components/typography/RadixLarge.svelte";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { useQueryClient } from "@tanstack/svelte-query";

  const runtimeClient = useRuntimeClient();
  $: organization = $page.params.organization;
  $: project = $page.params.project;

  const PAGE_SIZE = 50;

  $: magicAuthTokensInfiniteQuery =
    createAdminServiceListMagicAuthTokensInfinite(
      organization,
      project,
      { pageSize: PAGE_SIZE },
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

  // Auto-fetch all remaining pages
  $: if (
    $magicAuthTokensInfiniteQuery.hasNextPage &&
    !$magicAuthTokensInfiniteQuery.isFetchingNextPage
  ) {
    void $magicAuthTokensInfiniteQuery.fetchNextPage();
  }

  $: allRows =
    $magicAuthTokensInfiniteQuery.data?.pages.flatMap(
      (page) => page.tokens ?? [],
    ) ?? [];

  $: dashboards = useDashboards(runtimeClient);

  $: allRowsWithDashboardTitle = allRows.map((token) => {
    const dashboard = $dashboards.data?.find(
      (d) => d.meta?.name?.name === token.resources?.[0]?.name,
    );

    return {
      ...token,
      dashboardTitle:
        dashboard?.explore?.spec?.displayName ||
        dashboard?.canvas?.spec?.displayName ||
        dashboard?.meta?.name?.name ||
        "",
    };
  });

  // function useValidDashboardTitle(dashboard: V1Resource) {
  //   return (
  //     dashboard?.explore?.spec?.displayName ||
  //     dashboard?.canvas?.spec?.displayName ||
  //     dashboard?.meta.name.name
  //   );
  // }

  // REVISIT when server-side sorting is implemented
  $: sortedAllRowsWithDashboardTitle = allRowsWithDashboardTitle.sort(
    (a, b) =>
      new Date(b.createdOn ?? 0).getTime() -
      new Date(a.createdOn ?? 0).getTime(),
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

<div class="flex flex-col items-center gap-y-4 w-full">
  {#if $magicAuthTokensInfiniteQuery.isLoading}
    <div class="m-auto mt-20">
      <DelayedSpinner
        isLoading={$magicAuthTokensInfiniteQuery.isLoading}
        size="24px"
      />
    </div>
  {:else if $magicAuthTokensInfiniteQuery.isError}
    <p class="text-red-500">Error loading public URLs</p>
  {:else}
    <div class="flex flex-col gap-3 w-full overflow-hidden">
      <div class="flex flex-col">
        <RadixLarge>Public URLs</RadixLarge>
        <p class="text-sm text-fg-tertiary font-medium">
          Manage shared public URLs for your dashboards. <a
            href="https://docs.rilldata.com/guide/dashboards/public-urls"
            target="_blank"
            class="text-primary-600 hover:text-primary-700 active:text-primary-800"
          >
            Learn more ->
          </a>
        </p>
      </div>
      <PublicURLsResourceTable
        data={sortedAllRowsWithDashboardTitle}
        onDelete={handleDelete}
      />
    </div>
  {/if}
</div>
