<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import { onMount } from "svelte";
  import { ADMIN_URL } from "../../../../client/http-client";

  let user;
  let redirectURL;

  async function init() {
    const urlParams = new URLSearchParams(window.location.search);
    redirectURL = urlParams.get("redirect_url");

    const response = await fetch(ADMIN_URL + "/v1/users/current", {
      method: "GET",
      credentials: "include",
    });
    let data = await response.json();
    if (!data.user) {
      window.location.href =
        ADMIN_URL + "/auth/login?redirect=" + window.location.href;
    } else {
      user = data.user;
    }
  }
  function handleGoToGithub() {
    window.location.href = redirectURL;
  }
  onMount(init);
</script>

<svelte:head>
  <title>Connect to Github</title>
</svelte:head>

{#if user}
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
