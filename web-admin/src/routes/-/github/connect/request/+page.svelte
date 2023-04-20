<script lang="ts">
  import { goto } from "$app/navigation";
  import { ADMIN_URL } from "@rilldata/web-admin/client/http-client";
  import { createAdminServiceGetCurrentUser } from "@rilldata/web-admin/client";

  const remote = new URLSearchParams(window.location.search).get("remote");
  const user = createAdminServiceGetCurrentUser({
    query: {
      onSuccess: (data) => {
        if (!data.user) {
          goto(`${ADMIN_URL}/auth/login?redirect=${window.location.href}`);
        }
      },
    },
  });
</script>

<svelte:head>
  <title>Github access requested</title>
</svelte:head>

{#if $user.data && $user.data.user}
  <div class="flex flex-col justify-center items-center h-3/5">
    <h1 class="text-3xl font-medium text-gray-800 mb-4">Connect to Github</h1>
    <p class="text-lg text-gray-700 mb-6">
      You requested access to {@html remote}. You can close this page now.<br />
      CLI will keep polling until access has been granted by admin.<br />
      You can stop polling by pressing `ctrl+c` and run `rill deploy` again once
      access has been granted.
    </p>
  </div>
{/if}
