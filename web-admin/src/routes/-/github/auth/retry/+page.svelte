<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import { onMount } from "svelte";
  import { ADMIN_URL } from "../../../../../client/http-client";

  let remote;
  let githubUsername;
  let user

  async function init() {
    const urlParams = new URLSearchParams(window.location.search);
    remote = urlParams.get("remote");
    githubUsername = urlParams.get("githubUsername");
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
      user = data.user
    }
  }

  function handleGoToGithub() {
    window.location.href = encodeURI(
      ADMIN_URL + "/github/auth/login?remote=" + remote
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
      Your authorised github user {@html githubUsername} is not a collaborator to repo {@html remote}.<br>
      Click the button below to re-authorise/authorise another account.
    </p>
    <div class="mt-4">
      <Button type="primary" on:click={handleGoToGithub}>Go to Github</Button>
    </div>
  </div>
{/if}
