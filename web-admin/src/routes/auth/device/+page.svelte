<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
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

  function confirmUserCode() {
    fetch(
      ADMIN_URL +
        `/auth/oauth/device?user_code=${userCode}&code_confirmed=true`,
      {
        method: "POST",
        credentials: "include",
      }
    ).then((response) => {
      if (response.ok) {
        successMsg = "User code confirmed, this page can be closed now";
      } else {
        errorMsg = "User code confirmation failed";
        response.body
          .getReader()
          .read()
          .then(({ value }) => {
            const decoder = new TextDecoder("utf-8");
            errorMsg = errorMsg + ": " + decoder.decode(value);
          });
      }
    });
    actionTaken = true;
  }

  function rejectUserCode() {
    fetch(
      ADMIN_URL +
        `/auth/oauth/device?user_code=${userCode}&code_confirmed=false`,
      {
        method: "POST",
        credentials: "include",
      }
    ).then((response) => {
      if (response.ok) {
        successMsg = "User code rejected, this page can be closed now";
      } else {
        errorMsg = "User code rejection failed";
        response.body
          .getReader()
          .read()
          .then(({ value }) => {
            const decoder = new TextDecoder("utf-8");
            errorMsg = errorMsg + ": " + decoder.decode(value);
          });
      }
    });
  }

  onMount(fetchUserData);
</script>

<svelte:head>
  <meta name="description" content="User code confirmation" />
</svelte:head>

{#if user}
  <div class="flex flex-col justify-center items-center h-3/5">
    <h1 class="text-3xl font-medium text-gray-800 mb-4">
      Hello, {user.displayName}!
    </h1>
    <p class="text-lg text-gray-700 mb-6">Your user code is: {userCode}</p>

    <Button type="primary" on:click={() => {
        actionTaken = true;
        confirmUserCode();
      }}
      disabled={actionTaken}>Confirm</Button>
<div class="mt-4"></div>
      <Button type="secondary" on:click={() => {
        actionTaken = true;
        rejectUserCode();
      }}
      disabled={actionTaken}>Reject</Button>

    <div class="mt-4"></div>
    <p class="text-md text-green-700 font-bold mb-6">{successMsg}</p>
    <p class="text-md text-red-400 font-bold mb-6">{errorMsg}</p>
  </div>
{/if}
