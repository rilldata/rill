<script lang="ts">
  import { ADMIN_URL } from "@rilldata/web-admin/client/http-client";
  import CtaButton from "@rilldata/web-common/components/calls-to-action/CTAButton.svelte";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import RillLogoSquareNegative from "@rilldata/web-common/components/icons/RillLogoSquareNegative.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import type { PageData } from "./$types";

  export let data: PageData;

  $: ({ user } = data);

  let actionTaken = false;
  let successMsg = "";
  let errorMsg = "";
  const urlParams = new URLSearchParams(window.location.search);
  const redirectURL = urlParams.get("redirect");
  const userCode = urlParams.get("user_code");

  function confirmUserCode() {
    fetch(
      ADMIN_URL +
        `/auth/oauth/device?user_code=${userCode}&code_confirmed=true`,
      {
        method: "POST",
        credentials: "include",
      },
    ).then((response) => {
      if (response.ok) {
        if (redirectURL && redirectURL !== "") {
          window.location.href = decodeURIComponent(redirectURL);
        } else {
          successMsg = m.auth_device_code_confirmed();
        }
      } else {
        errorMsg = m.auth_device_code_confirmation_failed();
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
      },
    ).then((response) => {
      if (response.ok) {
        errorMsg = m.auth_device_code_rejected();
      } else {
        errorMsg = m.auth_device_code_rejection_failed();
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
  <meta name="description" content={m.auth_device_meta_description()} />
</svelte:head>

<CtaLayoutContainer>
  <CtaContentContainer>
    <RillLogoSquareNegative size="84px" />
    <h1 class="text-xl font-normal text-fg-primary">{m.auth_authorize_rill_cli()}</h1>
    <p class="text-base text-fg-secondary text-center">
      {m.auth_authenticating_as({ email: user.email })}<br />{m.auth_confirm_code_displayed()}
    </p>
    <div
      class="px-2 py-1 rounded-sm text-4xl tracking-widest bg-gray-100 text-fg-primary mb-5 font-mono"
    >
      {userCode}
    </div>

    <div class="flex flex-col gap-y-4 w-[400px]">
      <CtaButton
        variant="primary"
        onClick={() => {
          actionTaken = true;
          confirmUserCode();
        }}
        disabled={actionTaken}>{m.auth_confirm_code()}</CtaButton
      >
      <CtaButton
        variant="secondary"
        onClick={() => {
          actionTaken = true;
          rejectUserCode();
        }}
        disabled={actionTaken}>{m.common_cancel()}</CtaButton
      >
    </div>

    {#if successMsg}
      <p class="text-md text-green-700 font-bold mb-6">{successMsg}</p>
    {/if}
    {#if errorMsg}
      <p class="text-md text-red-400 font-bold mb-6">{errorMsg}</p>
    {/if}
  </CtaContentContainer>
</CtaLayoutContainer>
