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
  import type { V1MagicAuthToken } from "@rilldata/web-admin/client";

  $: organization = $page.params.organization;
  $: project = $page.params.project;

  let pageSize = 10;
  let pageToken: string | undefined = undefined;
  let allTokens = new Set<V1MagicAuthToken>();

  $: magicAuthTokens = createAdminServiceListMagicAuthTokens(
    organization,
    project,
    {
      pageSize,
      pageToken,
    },
  );

  $: if ($magicAuthTokens.data) {
    allTokens = new Set([...allTokens, ...$magicAuthTokens.data.tokens]);
  }

  const queryClient = useQueryClient();
  const revokeMagicAuthToken = createAdminServiceRevokeMagicAuthToken();

  async function handleDelete(deletedTokenId: string) {
    try {
      await $revokeMagicAuthToken.mutateAsync({
        tokenId: deletedTokenId,
      });

      // Optimistically update the local cache
      queryClient.setQueryData(
        getAdminServiceListMagicAuthTokensQueryKey(organization, project),
        (oldData: any) => ({
          ...oldData,
          tokens: oldData.tokens.filter(
            (token: any) => token.id !== deletedTokenId,
          ),
        }),
      );

      await queryClient.invalidateQueries(
        getAdminServiceListMagicAuthTokensQueryKey(organization, project),
      );

      eventBus.emit("notification", { message: "Public URL deleted" });
    } catch (error) {
      eventBus.emit("notification", {
        message: "Failed to delete public URL",
        type: "error",
      });
    }
  }

  // forward cursor-based pagination
  function handleLoadMore() {
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
            magicAuthTokens={Array.from(allTokens)}
            {pageSize}
            onDelete={handleDelete}
            onLoadMore={handleLoadMore}
            hasNextPage={!!$magicAuthTokens.data?.nextPageToken}
          />
        {/if}
      {/if}
    </div>
  </div>
</div>
