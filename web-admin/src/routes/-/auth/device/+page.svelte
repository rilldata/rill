<script lang="ts">
  import { page } from "$app/state";
  import { ADMIN_URL } from "@rilldata/web-admin/client/http-client";
  import { createDeviceAuthorizationAction } from "./device-authorization-action";
  import CtaButton from "@rilldata/web-common/components/calls-to-action/CTAButton.svelte";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import RillLogoSquareNegative from "@rilldata/web-common/components/icons/RillLogoSquareNegative.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import type { PageData } from "./$types";

  export let data: PageData;

  type DevicePageData = { user?: { email?: string | null } };
  $: userEmail = (data as DevicePageData).user?.email ?? "";

  let inFlight = false;
  let completed = false;
  let successMsg = "";
  let errorMsg = "";
  $: redirectURL = page.url.searchParams.get("redirect");
  $: userCode = page.url.searchParams.get("user_code");

  const authorizeDevice = createDeviceAuthorizationAction({
    adminUrl: ADMIN_URL as string,
    fetch: (input, init) => fetch(input, init),
    navigate: (url) => window.location.assign(url),
    messages: {
      confirmed: m.auth_device_code_confirmed(),
      rejected: m.auth_device_code_rejected(),
      confirmationFailed: m.auth_device_code_confirmation_failed(),
      rejectionFailed: m.auth_device_code_rejection_failed(),
      invalidCode: `${m.auth_device_code_confirmation_failed()}: Invalid user code`,
      unsafeRedirect: "The requested redirect destination is not allowed.",
    },
    onState: (state) => {
      inFlight = state.inFlight;
      completed = state.completed;
      successMsg = state.successMessage;
      errorMsg = state.errorMessage;
    },
  });

  function submit(confirmed: boolean) {
    void authorizeDevice({
      userCode,
      confirmed,
      redirect: redirectURL,
      currentUrl: page.url.toString(),
    });
  }
</script>

<svelte:head>
  <meta name="description" content={m.auth_device_meta_description()} />
</svelte:head>

<CtaLayoutContainer>
  <CtaContentContainer>
    <RillLogoSquareNegative size="84px" />
    <h1 class="text-xl font-normal text-fg-primary">
      {m.auth_authorize_rill_cli()}
    </h1>
    <p class="text-base text-fg-secondary text-center">
      {m.auth_authenticating_as({ email: userEmail })}<br
      />{m.auth_confirm_code_displayed()}
    </p>
    <div
      class="px-2 py-1 rounded-sm text-4xl tracking-widest bg-gray-100 text-fg-primary mb-5 font-mono"
    >
      {userCode}
    </div>

    <div class="flex flex-col gap-y-4 w-[400px]">
      <CtaButton
        variant="primary"
        onClick={() => submit(true)}
        disabled={inFlight || completed}>{m.auth_confirm_code()}</CtaButton
      >
      <CtaButton
        variant="secondary"
        onClick={() => submit(false)}
        disabled={inFlight || completed}>{m.common_cancel()}</CtaButton
      >
    </div>

    {#if successMsg}
      <p class="text-md text-green-700 font-bold mb-6">{successMsg}</p>
    {/if}
    {#if errorMsg}
      <p class="text-md text-red-400 font-bold mb-6" role="alert">
        {errorMsg}
      </p>
    {/if}
  </CtaContentContainer>
</CtaLayoutContainer>
