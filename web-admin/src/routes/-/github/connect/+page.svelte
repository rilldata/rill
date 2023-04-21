<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import { ADMIN_URL } from "@rilldata/web-admin/client/http-client";
  import { goto } from "$app/navigation";
  import { createAdminServiceGetCurrentUser } from "@rilldata/web-admin/client";

  const urlParams = new URLSearchParams(window.location.search);
  const redirectURL = urlParams.get("redirect");
  const remote = new URL(decodeURIComponent(redirectURL)).searchParams.get(
    "remote"
  );
  const user = createAdminServiceGetCurrentUser({
    query: {
      onSuccess: (data) => {
        if (!data.user) {
          goto(`${ADMIN_URL}/auth/login?redirect=${window.location.href}`);
        }
      },
    },
  });

  function handleGoToGithub() {
    window.location.href = redirectURL;
  }
</script>

<svelte:head>
  <title>Connect to Github</title>
</svelte:head>

{#if $user.data && $user.data.user}
  <div class="flex flex-col justify-center items-center h-3/5">
    <h1 class="text-3xl font-medium text-gray-800 mb-4">Connect to Github</h1>
    <p class="text-lg text-gray-700 mb-6">
      Authentication successful. Rill projects deploy continuously when you push
      changes to Github. <br />
      You need to grant Rill read only access to your repository `{@html remote}`
      on Github.
    </p>
    <div class="mt-4">
      <Button type="primary" on:click={handleGoToGithub}>Go to Github</Button>
    </div>
  </div>
{/if}
