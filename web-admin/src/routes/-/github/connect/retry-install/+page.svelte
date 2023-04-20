<!-- When we navigate users to install page. 
  We can't control the repo users install the github app on and they can end up installing the app on another repo.
  This page is for showing them the message that github app is installed on another repo than they need to reinstall app on right repo.  -->
<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
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
  function handleGoToGithub() {
    window.location.href = encodeURI(
      ADMIN_URL + "/github/connect?remote=" + remote
    );
  }
</script>

<svelte:head>
  <title>Connect to Github</title>
</svelte:head>

{#if $user.data && $user.data.user}
  <div class="flex flex-col justify-center items-center h-3/5">
    <h1 class="text-3xl font-medium text-gray-800 mb-4">Connect to Github</h1>
    <p class="text-lg text-gray-700 text-2xl mb-4">
      It looks like you did not grant access the the desired repository at {@html remote}.<br
      />
      Click the button below to retry (If this was intentional, press ctrl+c in the
      CLI to cancel the connect request)
    </p>
    <div class="mt-4">
      <Button type="primary" on:click={handleGoToGithub}>Go to Github</Button>
    </div>
  </div>
{/if}
