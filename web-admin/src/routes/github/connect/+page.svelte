<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import { onMount } from "svelte";
  import { ADMIN_URL } from "../../../client/http-client";

  let user;
  // let actionTaken = false;
  // let successMsg = "";
  // let errorMsg = "";

  async function init() {
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

    const urlParams = new URLSearchParams(window.location.search);
    let autoredirect = urlParams.get("auto_redirect");
    if (autoredirect === "true") {
      handleGoToGithub();
    }
  }

  function handleGoToGithub() {
    window.location.href = ADMIN_URL + "/github/connect";
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
      Click the button to grant Rill access to one or more of your Github
      repositories.
    </p>
    <div class="mt-4">
      <Button type="primary" on:click={handleGoToGithub}>Go to Github</Button>
    </div>
  </div>
{/if}
