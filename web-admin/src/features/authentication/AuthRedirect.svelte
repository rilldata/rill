<script lang="ts">
  import { goto } from "$app/navigation";
  import { createAdminServiceGetCurrentUser } from "../../client";
  import { CANONICAL_ADMIN_URL } from "../../client/http-client";

  const user = createAdminServiceGetCurrentUser();

  // redirect to login if not logged in
  $: if ($user.isSuccess && !$user.data.user) {
    goto(`${CANONICAL_ADMIN_URL}/auth/login?redirect=${window.origin}`);
  }
</script>

{#if $user.data && $user.data.user}
  <slot />
{/if}
