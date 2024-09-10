<script lang="ts">
  import PublicURLsTable from "@rilldata/web-admin/features/public-urls/PublicURLsTable.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
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

  $: organization = $page.params.organization;
  $: project = $page.params.project;
  $: magicAuthTokensQuery = createAdminServiceListMagicAuthTokens(
    organization,
    project,
  );

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
</script>

<div class="flex flex-col w-full">
  <div class="flex md:flex-row flex-col gap-6">
    <div class="w-full">
      {#if $magicAuthTokensQuery.isLoading}
        <DelayedSpinner
          isLoading={$magicAuthTokensQuery.isLoading}
          size="1rem"
        />
      {:else if $magicAuthTokensQuery.error}
        <div class="text-red-500">
          Error loading resources: {$magicAuthTokensQuery.error?.message}
        </div>
      {:else if $magicAuthTokensQuery.data}
        {#if $magicAuthTokensQuery.data.tokens.length === 0}
          <NoPublicURLCTA />
        {:else}
          <PublicURLsTable
            magicAuthTokens={$magicAuthTokensQuery.data.tokens}
            onDelete={handleDelete}
          />
        {/if}
      {/if}
    </div>
  </div>
</div>
