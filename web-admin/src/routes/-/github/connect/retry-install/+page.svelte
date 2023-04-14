<!-- When we navigate users to install page. 
  We can't control the repo users install the github app on and they can end up installing the app on another repo.
  This page is for showing them the message that github app is installed on another repo than they need to reinstall app on right repo.  -->
<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import { onMount } from "svelte";
  import { ADMIN_URL } from "../../../../../client/http-client";
  let remote;
  let user;
  async function init() {
    const urlParams = new URLSearchParams(window.location.search);
    remote = urlParams.get("remote");
    const response = await fetch(ADMIN_URL + "/v1/users/current", {
      method: "GET",
      credentials: "include",
    });
    let data = await response.json();
    if (!data.user) {
      // this should not happen since user is already authenticated
      window.location.href =
        ADMIN_URL + "/auth/login?redirect=" + window.location.href;
    } else {
      user = data.user;
    }
  }
  function handleGoToGithub() {
    window.location.href = encodeURI(
      ADMIN_URL + "/github/connect?remote=" + remote
    );
  }
  onMount(init);
</script>

<svelte:head>
  <title>Connect to Github</title>
</svelte:head>

{#if user}
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
