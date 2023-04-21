<script lang="ts">
  import RillLogoSquareNegative from "@rilldata/web-common/components/icons/RillLogoSquareNegative.svelte";
  import { onMount } from "svelte";
  import type { V1User } from "../../../../client";
  import { ADMIN_URL } from "../../../../client/http-client";
  import CtaButton from "../../../../components/CTAButton.svelte";

  let user: V1User;
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
  <div class="flex flex-col justify-center items-center h-4/5 gap-y-6">
    <RillLogoSquareNegative size="84px" />
    <h1 class="text-xl font-normal text-gray-800">Authorize Rill CLI</h1>
    <p class="text-base text-gray-500 text-center">
      You are authenticating into Rill as <span
        class="font-medium text-gray-600">{user.email}</span
      >.<br />Please confirm this is the code displayed in the Rill CLI.
    </p>
    <div
      class="px-2 py-1 rounded-sm text-4xl tracking-widest bg-gray-100 text-gray-700 mb-5 font-mono"
    >
      {userCode}
    </div>

    <div class="flex flex-col gap-y-4 w-[400px]">
      <CtaButton
        variant="primary"
        on:click={() => {
          actionTaken = true;
          confirmUserCode();
        }}
        disabled={actionTaken}>Confirm code</CtaButton
      >
      <CtaButton
        variant="secondary"
        on:click={() => {
          actionTaken = true;
          rejectUserCode();
        }}
        disabled={actionTaken}>Cancel</CtaButton
      >
    </div>

    {#if successMsg}
      <p class="text-md text-green-700 font-bold mb-6">{successMsg}</p>
    {/if}
    {#if errorMsg}
      <p class="text-md text-red-400 font-bold mb-6">{errorMsg}</p>
    {/if}
  </div>
{/if}
