<script lang="ts">
  import { onMount } from "svelte";
  import { ADMIN_URL } from "../../../client/http-client";

  let user;
  let userCode;
  let actionTaken = false;
  let successMsg = "";
  let errorMsg = "";

  async function fetchUserData() {
    const urlParams = new URLSearchParams(window.location.search);
    userCode = urlParams.get("user_code");

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

  onMount(fetchUserData);
</script>

<svelte:head>
  <title>Connect to Github</title>
</svelte:head>

{#if user}
  <section>
    <h2>Connect to Github</h2>
    <p>
      Click the button to grant Rill access to one or more of your Github
      repositories.
    </p>
    <a href="{ADMIN_URL}/github/connect">Go to Github</a>
  </section>
{/if}
