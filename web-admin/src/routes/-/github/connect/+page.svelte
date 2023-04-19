<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import { ADMIN_URL } from "../../../../client/http-client";
  import { goto } from "$app/navigation";
  import { createAdminServiceGetCurrentUser } from "../../../../client";

  const redirectURL = new URLSearchParams(window.location.search).get(
    "redirect_url"
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
      You need to grant Rill read only access to your repository on Github.
    </p>
    <div class="mt-4">
      <Button type="primary" on:click={handleGoToGithub}>Go to Github</Button>
    </div>
  </div>
{/if}
