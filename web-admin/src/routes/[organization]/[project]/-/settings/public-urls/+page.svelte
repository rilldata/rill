<script lang="ts">
  import PublicURLsTable from "@rilldata/web-admin/features/public-urls/PublicURLsTable.svelte";
  import { page } from "$app/stores";
  import {
    createAdminServiceListMagicAuthTokens,
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
  import {
    adminServiceListMagicAuthTokens,
    createAdminServiceListMagicAuthTokensInfiniteQuery,
  } from "@rilldata/web-admin/features/public-urls/create-infinite-query-public-urls";

  $: organization = $page.params.organization;
  $: project = $page.params.project;

  $: infiniteSplitsQuery = createAdminServiceListMagicAuthTokensInfiniteQuery(
    organization,
    project,
  );

  function useValidDashboardTitle(dashboard: DashboardResource) {
    return (
      dashboard?.resource.metricsView.state?.validSpec?.title ||
      dashboard?.resource.meta.name.name
    );
  }

  $: allRows =
    $infiniteSplitsQuery.data?.pages.flatMap((page) => page.tokens ?? []) ?? [];

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

  const queryClient = useQueryClient();
  const revokeMagicAuthToken = createAdminServiceRevokeMagicAuthToken();

  async function handleDelete(deletedTokenId: string) {
    try {
      // Perform the deletion
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
      {#if $infiniteSplitsQuery.isLoading}
        <DelayedSpinner
          isLoading={$infiniteSplitsQuery.isLoading}
          size="1rem"
        />
      {:else if $infiniteSplitsQuery.error}
        <div class="text-red-500">
          Error loading public URLs: {$infiniteSplitsQuery.error}
        </div>
      {:else if !$infiniteSplitsQuery.data.pages.length}
        <NoPublicURLCTA />
      {:else}
        <PublicURLsTable
          data={allRowsWithDashboardTitle}
          onDelete={handleDelete}
          query={$infiniteSplitsQuery}
        />
      {/if}
    </div>
  </div>
</div>
