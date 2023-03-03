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

    const response = await fetch(ADMIN_URL + "/auth/user", {
      method: "GET",
      credentials: "include",
    });
    console.log(response.status);
    if (response.status === 401) {
      window.location.href =
        ADMIN_URL + "/auth/login?redirect=" + window.location.href;
    } else if (response.ok) {
      user = await response.json();
    }
  }

  function confirmUserCode() {
    fetch(
      ADMIN_URL + `/oauth/device?user_code=${userCode}&code_confirmed=true`,
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
      ADMIN_URL + `/oauth/device?user_code=${userCode}&code_confirmed=false`,
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
  <div>
    <h1>Hello, {user.name}!</h1>
    <p>Your user code is: {userCode}</p>
    <button
      on:click={() => {
        actionTaken = true;
        confirmUserCode();
      }}
      disabled={actionTaken}>Confirm</button
    >
    <button
      on:click={() => {
        actionTaken = true;
        rejectUserCode();
      }}
      disabled={actionTaken}>Reject</button
    >
    <p style="color: green; font-weight: bold">{successMsg}</p>
    <p style="color: red; font-weight: bold">{errorMsg}</p>
  </div>
{/if}
