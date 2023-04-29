<script lang="ts">
  import HomeShareCTA from "@rilldata/web-admin/components/home/HomeShareCTA.svelte";
  import {
    createAdminServiceGetCurrentUser,
    createAdminServiceListOrganizations,
  } from "../client";
  import AuthRedirect from "../components/authentication/AuthRedirect.svelte";
  import OrganizationList from "../components/home/OrganizationList.svelte";
  import WelcomeMessage from "../components/home/WelcomeMessage.svelte";

  const user = createAdminServiceGetCurrentUser();

  const orgs = createAdminServiceListOrganizations(undefined, {
    query: {
      placeholderData: undefined,
    },
  });
</script>

<svelte:head>
  <title>Home - Rill</title>
</svelte:head>

<AuthRedirect>
  <section class="flex flex-col w-4/5 mx-auto h-5/6 my-36">
    <h1 class="text-4xl leading-10 font-light mb-2">
      Hi {$user.data.user.displayName}!
    </h1>
    <div class="flex flex-row">
      <div class="w-1/2">
        {#if $orgs.isSuccess}
          {#if $orgs.data.organizations.length === 0}
            <WelcomeMessage />
          {:else}
            <h3 class="text-base leading-6 font-normal text-gray-500 mb-2">
              Check out your dashboards below.
            </h3>
            <OrganizationList />
          {/if}
        {/if}
      </div>
      <HomeShareCTA />
    </div>
  </section>
</AuthRedirect>
