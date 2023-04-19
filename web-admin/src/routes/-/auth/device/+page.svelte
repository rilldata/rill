<script lang="ts">
  import { goto } from "$app/navigation";
  import { createAdminServiceGetCurrentUser } from "../../../../client";
  import { Button } from "@rilldata/web-common/components/button";
  import { ADMIN_URL } from "../../../../client/http-client";

  let actionTaken = false;
  let successMsg = "";
  let errorMsg = "";
  const urlParams = new URLSearchParams(window.location.search)
  const redirectURL = urlParams.get("redirect_url");
  const userCode = urlParams.get("user_code");

  const user = createAdminServiceGetCurrentUser({
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
          // This msg creates a tight coupling b/w login flow and github access flow
          // but is required for better user experience
          successMsg = "User code confirmed, verifying github access...";
          setTimeout(function () {
            window.location.href = decodeURIComponent(redirectURL);
          }, 2 * 1000);
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
  <div class="flex flex-col justify-center items-center h-3/5">
    <h1 class="text-3xl font-medium text-gray-800 mb-4">
      Hello, {$user.data.user.displayName}!
    </h1>
    <p class="text-lg text-gray-700 mb-6">Your user code is: {userCode}</p>

    <Button
      type="primary"
      on:click={() => {
        actionTaken = true;
        confirmUserCode();
      }}
      disabled={actionTaken}>Confirm</Button
    >
    <div class="mt-4" />
    <Button
      type="secondary"
      on:click={() => {
        actionTaken = true;
        rejectUserCode();
      }}
      disabled={actionTaken}>Reject</Button
    >

    <div class="mt-4" />
    <p class="text-md text-green-700 font-bold mb-6">{successMsg}</p>
    <p class="text-md text-red-400 font-bold mb-6">{errorMsg}</p>
  </div>
{/if}
