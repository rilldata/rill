<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import { onMount } from "svelte";
  import { ADMIN_URL } from "../../../../client/http-client";
  let remote;
  let user;

  async function fetchUserData() {
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

  onMount(fetchUserData);
</script>

<svelte:head>
  <title>Github access requested</title>
</svelte:head>

{#if user}
  <div class="flex flex-col justify-center items-center h-3/5">
    <h1 class="text-3xl font-medium text-gray-800 mb-4">Connect to Github</h1>
    <p class="text-lg text-gray-700 mb-6">
      You requested access to {@html remote}. You can close this page now and
      continue once access has been granted.
    </p>
  </div>
{/if}
