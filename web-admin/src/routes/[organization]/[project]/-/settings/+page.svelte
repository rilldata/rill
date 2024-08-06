<script lang="ts">
  import ContentContainer from "@rilldata/web-admin/components/layout/ContentContainer.svelte";
  import PublicURLsTable from "@rilldata/web-admin/features/public-urls/PublicURLsTable.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { page } from "$app/stores";
  import { createAdminServiceListMagicAuthTokens } from "@rilldata/web-admin/client";
  import { onMount } from "svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;

  $: magicAuthTokensQuery = createAdminServiceListMagicAuthTokens(
    organization,
    project,
  );
  $: magicAuthTokens = $magicAuthTokensQuery.data?.tokens ?? [];

  $: settingsNavs = [
    {
      label: "Public URLs",
      hash: "#public-urls",
    },
  ];

  onMount(() => {
    const defaultNav = settingsNavs.find((nav) => nav.hash === "#public-urls");
    if (!$page.url.hash && defaultNav) {
      window.location.hash = defaultNav.hash;
    }
  });
</script>

<ContentContainer>
  <div class="flex flex-col w-full">
    <h3 class="text-lg font-medium text-slate-700">Settings</h3>

    <div class="mt-6 flex md:flex-row flex-col gap-6">
      <aside class="w-full md:w-1/4 flex flex-col gap-2">
        {#each settingsNavs as nav (nav.hash)}
          <a
            href={`${nav.hash}`}
            class="hover:bg-slate-100 rounded p-2 w-full"
            class:bg-slate-100={$page.url.hash === nav.hash}
            class:active={$page.url.hash === nav.hash}
          >
            <h3 class="text-md font-medium text-slate-900">
              {nav.label}
            </h3>
          </a>
        {/each}
      </aside>

      <div class="w-full md:w-3/4">
        {#if $page.url.hash === "#public-urls"}
          {#if $magicAuthTokensQuery.isLoading}
            <Spinner status={EntityStatus.Running} size={"16px"} />
          {:else if $magicAuthTokensQuery.error}
            <div class="text-red-500">
              Error loading resources: {$magicAuthTokensQuery.error?.message}
            </div>
          {:else if $magicAuthTokensQuery.data}
            <PublicURLsTable {magicAuthTokens} {organization} {project} />
          {/if}
        {/if}
      </div>
    </div>
  </div>
</ContentContainer>
