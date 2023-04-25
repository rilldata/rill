<script lang="ts">
  import { createAdminServiceGetCurrentUser } from "../client";
  import AuthRedirect from "../components/authentication/AuthRedirect.svelte";
  import OrganizationList from "../components/home/OrganizationList.svelte";

  const user = createAdminServiceGetCurrentUser();
</script>

<svelte:head>
  <title>Home - Rill</title>
</svelte:head>

<AuthRedirect>
  <section class="flex flex-col justify-center w-4/5 mx-auto h-2/5">
    <h1 class="text-4xl leading-10 font-light mb-2">
      Hi {$user.data.user.displayName}!
    </h1>
    <h3 class="text-base leading-6 font-normal text-gray-500 mb-2">
      Check out your dashboards below.
    </h3>
    <OrganizationList />
  </section>
</AuthRedirect>
