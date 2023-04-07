<script lang="ts">
  import { useAdminServiceGetCurrentUser } from "../client";
  import OrganizationList from "../components/home/OrganizationList.svelte";

  const user = useAdminServiceGetCurrentUser();
  if ($user.data && !$user.data.user) {
    // throw redirect(307, `${ADMIN_URL}/auth/login?redirect=${window.origin}`);
  }
</script>

<section class="flex flex-col justify-center w-4/5 mx-auto h-2/5">
  <h1 class="text-4xl leading-10 font-light mb-2">
    {#if $user.data && $user.data.user}
      Hi {$user.data.user.displayName}!
    {/if}
  </h1>
  <h3 class="text-base leading-6 font-normal text-gray-500 mb-2">
    Check out your dashboards below.
  </h3>
  <OrganizationList />
</section>
