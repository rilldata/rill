<!-- This page is for cases when user authorised the github app on another github account which doesn't have access to the repo  -->
<script lang="ts">
  import { goto } from "$app/navigation";
  import { Button } from "@rilldata/web-common/components/button";
  import { ADMIN_URL } from "../../../../../client/http-client";
  import { createAdminServiceGetCurrentUser } from "../../../../../client";

  const urlParams = new URLSearchParams(window.location.search);
  const remote = urlParams.get("remote");
  const githubUsername = urlParams.get("githubUsername");
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
      ADMIN_URL + "/github/auth/login?remote=" + remote
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
      Your authorised github user {@html githubUsername} is not a collaborator to
      repo {@html remote}.<br />
      Click the button below to re-authorise/authorise another account.
    </p>
    <div class="mt-4">
      <Button type="primary" on:click={handleGoToGithub}>Go to Github</Button>
    </div>
  </div>
{/if}
