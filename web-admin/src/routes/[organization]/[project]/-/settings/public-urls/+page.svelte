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
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  $: organization = $page.params.organization;
  $: project = $page.params.project;

  let pageSize = 16;
  let pageToken: string | undefined = undefined;

  $: magicAuthTokens = createAdminServiceListMagicAuthTokens(
    organization,
    project,
    {
      pageSize,
      pageToken,
    },
  );

  $: dashboards = useDashboardsV2($runtime.instanceId);

  // Reset and update allTokensStore when magicAuthTokens changes
  $: currentPageTokens =
    $magicAuthTokens.data && $dashboards.data
      ? $magicAuthTokens.data.tokens.map((token) => {
          const dashboard = $dashboards.data.find(
            (d) => d.resource.meta.name.name === token.metricsView,
          );
          return {
            ...token,
            dashboardTitle:
              dashboard?.resource.metricsView.spec.title || "Unknown Dashboard",
          };
        })
      : [];

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

  // Forward cursor-based pagination with `nextPageToken`
  function handleNextPage() {
    if ($magicAuthTokens.data?.nextPageToken) {
      pageToken = $magicAuthTokens.data.nextPageToken;
    }
  }
</script>

<div class="flex flex-col w-full">
  <div class="flex md:flex-row flex-col gap-6">
    <div class="w-full">
      {#if $magicAuthTokens.isLoading}
        <DelayedSpinner isLoading={$magicAuthTokens.isLoading} size="1rem" />
      {:else if $magicAuthTokens.error}
        <div class="text-red-500">
          Error loading resources: {$magicAuthTokens.error?.message}
        </div>
      {:else if $magicAuthTokens.data}
        {#if $magicAuthTokens.data.tokens.length === 0}
          <NoPublicURLCTA />
        {:else}
          <PublicURLsTable
            tableData={currentPageTokens}
            {pageSize}
            onDelete={handleDelete}
            onNextPage={handleNextPage}
            hasNextPage={!!$magicAuthTokens.data?.nextPageToken}
          />
        {/if}
      {/if}
    </div>
  </div>
</div>
