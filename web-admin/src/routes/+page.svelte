<script lang="ts">
  import HomeShareCTA from "@rilldata/web-admin/components/home/HomeShareCTA.svelte";
  import {
    createAdminServiceGetCurrentUser,
    createAdminServiceListOrganizations,
  } from "../client";
  import AuthRedirect from "../components/authentication/AuthRedirect.svelte";
  import OrganizationList from "../components/home/OrganizationList.svelte";
  import WelcomeMessage from "../components/home/WelcomeMessage.svelte";

  import VerticalScrollContainer from "@rilldata/web-common/layout/VerticalScrollContainer.svelte";

  const user = createAdminServiceGetCurrentUser();

  const orgs = createAdminServiceListOrganizations(undefined, {
    query: {
      placeholderData: undefined,
    },
  });

  $: hasAnOrganization = $orgs.data?.organizations?.length > 0;

  function getFirstNameFromDisplayName(displayName: string) {
    return displayName.split(" ")[0];
  }
</script>

<svelte:head>
  <title>Home - Rill</title>
</svelte:head>

<AuthRedirect>
  <VerticalScrollContainer>
    <section
      class="flex flex-col mx-8 my-8 sm:my-16 sm:mx-16 lg:mx-32 lg:my-24 2xl:mx-64 mx-auto"
    >
      <h1 class="text-4xl leading-10 font-light mb-2">
        Hi {getFirstNameFromDisplayName($user.data.user.displayName)}!
      </h1>
      <div class="flex flex-row gap-x-7 flex-wrap">
        <div class="md:w-1/2">
          {#if $orgs.isSuccess}
            {#if !hasAnOrganization}
              <WelcomeMessage />
            {:else}
              <h3 class="text-base leading-6 font-normal text-gray-500 mb-3">
                Check out your dashboards below.
              </h3>
              <OrganizationList />
            {/if}
          {/if}
        </div>
        {#if hasAnOrganization}
          <HomeShareCTA />
        {/if}
      </div>
    </section>
  </VerticalScrollContainer>
</AuthRedirect>
