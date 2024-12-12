<script lang="ts">
  import CtaButton from "@rilldata/web-common/components/calls-to-action/CTAButton.svelte";
  import { getConnectionFromEmail, validateEmail } from "./utils";
  import { createEventDispatcher } from "svelte";

  import { WebAuth } from "auth0-js";

  export let disabled = false;
  export let webAuth: WebAuth;
  export let connectionMapObj: Record<string, string[]>;

  let email = "";
  let errorText = "";

  const dispatch = createEventDispatcher();

  const inputClasses =
    "h-10 px-4 py-2 border border-slate-300 rounded-sm text-base";
  const focusClasses =
    "ring-offset-2 focus:ring-2 focus:ring-primary-ry-300 focus:outline-none";

  function handleContinueEmailClick() {
    if (!email) {
      errorText = "Please enter your email";
      return;
    }

    if (!validateEmail(email)) {
      errorText = "Please enter a valid email address";
      return;
    }

    errorText = "";

    dispatch("submit", { email });
  }

  function authorizeSSO(email: string, connectionName: string) {
    webAuth.authorize({
      connection: connectionName,
      login_hint: email,
      prompt: "login",
    });
  }

  function handleContinueSSOClick() {
    const connectionName = getConnectionFromEmail(email, connectionMapObj);

    if (!connectionName) {
      errorText = `IDP for the email ${email} not found. Please contact your administrator.`;
      return;
    }

    authorizeSSO(email, connectionName);
  }

  $: if (email) {
    errorText = "";
  }

  $: disabled = !(email.length > 0 && validateEmail(email));
</script>

<div class="mb-4 flex flex-col gap-y-4">
  <input
    class="{inputClasses} {focusClasses}"
    style:width="400px"
    type="email"
    placeholder="Enter your email address"
    id="email"
    bind:value={email}
  />

  {#if errorText}
    <div class="text-red-500 text-sm -mt-2">
      {errorText}
    </div>
  {/if}
</div>

<CtaButton {disabled} variant="secondary" on:click={handleContinueEmailClick}>
  <div class="flex justify-center font-medium">
    <span>Continue with email</span>
  </div>
</CtaButton>

<CtaButton {disabled} variant="secondary" on:click={handleContinueSSOClick}>
  <div class="flex justify-center font-medium">
    <div>Continue with SAML SSO</div>
  </div>
</CtaButton>
