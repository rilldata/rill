<script lang="ts">
  import { goto } from "$app/navigation";
  import { createAdminServiceGetCurrentUser } from "@rilldata/web-admin/client";
  import { Button } from "@rilldata/web-common/components/button";
  import { ADMIN_URL } from "@rilldata/web-admin/client/http-client";

  import RillLogoSquareNegative from "@rilldata/web-common/components/icons/RillLogoSquareNegative.svelte";
  import type { V1User } from "@rilldata/web-admin/client";
  import CtaButton from "@rilldata/web-admin/components/CTAButton.svelte";

  let actionTaken = false;
  let successMsg = "";
  let errorMsg = "";
  const urlParams = new URLSearchParams(window.location.search);
  const redirectURL = urlParams.get("redirect");
  const userCode = urlParams.get("user_code");
  const user: V1User = createAdminServiceGetCurrentUser({
    query: {
      onSuccess: (data) => {
        if (!data.user) {
          goto(`${ADMIN_URL}/auth/login?redirect=${window.location.href}`);
        }
      },
    },
  });

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
        if (redirectURL !== "") {
          window.location.href = decodeURIComponent(redirectURL);
        } else {
          successMsg = "User code confirmed, this page can be closed now";
        }
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
</script>

<svelte:head>
  <meta name="description" content="User code confirmation" />
</svelte:head>

{#if $user.data && $user.data.user}
  <div class="flex flex-col justify-center items-center h-4/5 gap-y-6">
    <RillLogoSquareNegative size="84px" />
    <h1 class="text-xl font-normal text-gray-800">Authorize Rill CLI</h1>
    <p class="text-base text-gray-500 text-center">
      You are authenticating into Rill as <span
        class="font-medium text-gray-600">{$user.data.user.email}</span
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
