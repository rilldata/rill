<script lang="ts">
  import ContentContainer from "@rilldata/web-admin/components/layout/ContentContainer.svelte";
  import PublicURLsTable from "@rilldata/web-admin/features/public-urls/PublicURLsTable.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { page } from "$app/stores";
  import { createAdminServiceListMagicAuthTokens } from "@rilldata/web-admin/client";

  $: organization = $page.params.organization;
  $: project = $page.params.project;

  $: magicAuthTokensQuery = createAdminServiceListMagicAuthTokens(
    organization,
    project,
  );
  $: magicAuthTokens = $magicAuthTokensQuery.data?.tokens ?? [];
</script>

<ContentContainer>
  <div class="flex flex-col w-full">
    <!-- TODO: what is the token for radix/h3? -->
    <!-- TODO: font color -->
    <h3 class="text-lg font-medium">Settings</h3>

    <!-- TODO: placeholder, to put this to the left sidebar -->
    <div class="mt-6">
      <h3 class="text-md font-medium">Public URLs</h3>
    </div>

    <div class="mt-6">
      {#if $magicAuthTokensQuery.isLoading}
        <Spinner status={EntityStatus.Running} size={"16px"} />
      {:else if $magicAuthTokensQuery.error}
        <div class="text-red-500">
          Error loading resources: {$magicAuthTokensQuery.error?.message}
        </div>
      {:else if $magicAuthTokensQuery.data}
        <PublicURLsTable {magicAuthTokens} {organization} {project} />
      {/if}
    </div>
  </div>
</ContentContainer>
