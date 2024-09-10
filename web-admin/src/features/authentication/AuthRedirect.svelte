<script lang="ts">
  import { redirectToLogin } from "@rilldata/web-admin/client/redirect-utils";
  import { createAdminServiceGetCurrentUser } from "../../client";

  const user = createAdminServiceGetCurrentUser();

  // redirect to login if not logged in
  $: if ($user.isSuccess && !$user.data.user) {
    redirectToLogin();
  }
</script>

{#if $user.data && $user.data.user}
  <slot />
{/if}
