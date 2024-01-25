<script lang="ts">
  import { createAdminServiceGetCurrentUser } from "../client";
  import AuthRedirect from "../features/authentication/AuthRedirect.svelte";
  import WelcomeMessage from "../features/home/WelcomeMessage.svelte";
  import OrganizationRedirect from "../features/organizations/OrganizationRedirect.svelte";

  const user = createAdminServiceGetCurrentUser();

  function getFirstNameFromDisplayName(displayName: string) {
    return displayName.split(" ")[0];
  }
</script>

<svelte:head>
  <title>Home - Rill</title>
</svelte:head>

<AuthRedirect>
  <OrganizationRedirect>
    <section
      class="flex flex-col mx-8 my-8 sm:my-16 sm:mx-16 lg:mx-32 lg:my-24 2xl:mx-64 mx-auto"
    >
      <h1 class="text-4xl leading-10 font-light mb-2">
        Hi {getFirstNameFromDisplayName($user.data.user.displayName)}!
      </h1>
      <WelcomeMessage />
    </section>
  </OrganizationRedirect>
</AuthRedirect>
