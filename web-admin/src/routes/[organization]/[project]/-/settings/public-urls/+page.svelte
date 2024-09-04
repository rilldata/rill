<script lang="ts">
  import PublicURLsTable from "@rilldata/web-admin/features/public-urls/PublicURLsTable.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { page } from "$app/stores";
  import {
    createAdminServiceListMagicAuthTokens,
    getAdminServiceListMagicAuthTokensQueryKey,
    adminServiceRevokeMagicAuthToken,
  } from "@rilldata/web-admin/client";
  import NoPublicURLCTA from "@rilldata/web-admin/features/public-urls/NoPublicURLCTA.svelte";
  import { useQueryClient } from "@tanstack/svelte-query";

  const queryClient = useQueryClient();

  $: organization = $page.params.organization;
  $: project = $page.params.project;

  $: magicAuthTokensQuery = createAdminServiceListMagicAuthTokens(
    organization,
    project,
  );

  async function handleDelete(deletedTokenId: string) {
    await adminServiceRevokeMagicAuthToken(deletedTokenId);

    queryClient.refetchQueries(
      getAdminServiceListMagicAuthTokensQueryKey(organization, project),
    );
  }
</script>

<div class="flex flex-col w-full">
  <div class="flex md:flex-row flex-col gap-6">
    <div class="w-full">
      {#if $magicAuthTokensQuery.isLoading}
        <Spinner status={EntityStatus.Running} size={"16px"} />
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
            {organization}
            {project}
            onDelete={handleDelete}
          />
        {/if}
      {/if}
    </div>
  </div>
</div>
