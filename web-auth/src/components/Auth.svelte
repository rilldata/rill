<script lang="ts">
  import CtaButton from "@rilldata/web-common/components/calls-to-action/CTAButton.svelte";
  import RillLogoSquareNegative from "@rilldata/web-common/components/icons/RillLogoSquareNegative.svelte";
  import auth0, { WebAuth } from "auth0-js";
  import type { AuthOptions } from "auth0-js";
  import { onMount } from "svelte";
  import { LOGIN_OPTIONS } from "../config";
  import AuthContainer from "./AuthContainer.svelte";
  import EmailPasswordForm from "./EmailPasswordForm.svelte";
  import { getConnectionFromEmail } from "./utils";
  import OrSeparator from "./OrSeparator.svelte";
  import EmailSubmissionForm from "./EmailSubmissionForm.svelte";
  import Disclaimer from "./Disclaimer.svelte";
  import Spacer from "./Spacer.svelte";
  import { AuthStep, type Config } from "../types";
  import CtaNeedHelp from "@rilldata/web-common/components/calls-to-action/CTANeedHelp.svelte";

  export let configParams: string;
  export let cloudClientIDs = "";
  export let disableForgotPassDomains = "";
  export let connectionMap = "{}";

  const connectionMapObj = JSON.parse(connectionMap);
  const cloudClientIDsArr = cloudClientIDs.split(",");
  const disableForgotPassDomainsArr = disableForgotPassDomains.split(",");

  $: errorText = "";

  let email = "";
  let step: AuthStep = AuthStep.Base;
  let webAuth: WebAuth;

  // $: isLegacy = false;

  function isDomainDisabled(email: string): boolean {
    return disableForgotPassDomainsArr.some((domain) =>
      email.toLowerCase().endsWith(domain.toLowerCase()),
    );
  }

  $: domainDisabled = isDomainDisabled(email);

  function initConfig() {
    try {
      if (
        import.meta.env.DEV &&
        (!configParams || configParams === "undefined")
      ) {
        console.warn(
          "No auth config provided. In development mode - auth flows will not work.",
        );
        errorText = "Authentication is not configured in development mode";
        return;
      }

      const config = JSON.parse(
        decodeURIComponent(escape(window.atob(configParams))),
      ) as Config;

      const isSignup = config?.extraParams?.screen_hint === "signup";

      if (isSignup) {
        step = AuthStep.SignUp;
      }

      // UNCOMMENT to use `isLegacy` for rill dash check if needed
      // NOTE: Check for cloud client ids from auth0 to allow sign up and login
      // NOTE: Prevent rill dash (legacy) users from signing up
      // if (cloudClientIDsArr.includes(config?.clientID)) {
      //   isLegacy = true;
      // }

      const authOptions: AuthOptions = Object.assign(
        {
          overrides: {
            __tenant: config.auth0Tenant,
            __token_issuer: config.authorizationServer.issuer,
          },
          domain: config.auth0Domain,
          clientID: config.clientID,
          redirectUri: config.callbackURL,
          responseType: "code",
        },
        config.internalOptions,
      );

      webAuth = new auth0.WebAuth(authOptions);
    } catch (e) {
      console.error("Failed to initialize auth:", e);
      errorText = "Failed to initialize authentication in development mode";
    }
  }

  function authorizeSSO(email: string, connectionName: string) {
    webAuth.authorize({
      connection: connectionName,
      login_hint: email,
      prompt: "login",
    });
  }

  async function processEmailSubmission(event) {
    email = event.detail.email;
    const connectionName = getConnectionFromEmail(email, connectionMapObj);

    if (connectionName) {
      authorizeSSO(email, connectionName);
    } else {
      step = AuthStep.SignUp;
    }
  }

  function getHeadingText(step: AuthStep): string {
    switch (step) {
      case AuthStep.Base:
        return "Log in or sign up";
      case AuthStep.Login:
        return "Log in with email";
      case AuthStep.SignUp:
        return `Log in or sign up with <span class="font-medium">${email}</span>`;
      case AuthStep.Thanks:
        return "Thanks for signing up!";
      default:
        return "";
    }
  }

  function getSubheadingText(step: AuthStep, email: string): string {
    switch (step) {
      case AuthStep.Login:
        return `Log in using <span class="font-medium">${email}</span>`;
      default:
        return "";
    }
  }

  function backToBaseStep() {
    step = AuthStep.Base;
  }

  onMount(() => {
    initConfig();
  });

  $: headingText = getHeadingText(step);
  $: subheadingText = getSubheadingText(step, email);
</script>

<AuthContainer>
  <RillLogoSquareNegative size="84px" />
  <Spacer />
  <div class="flex flex-col items-center gap-y-2 text-center">
    <div class="text-xl text-slate-800">
      {@html headingText}
    </div>
    {#if subheadingText}
      <div class="text-base text-gray-500">
        {@html subheadingText}
      </div>
    {:else}
      <Spacer />
    {/if}
  </div>

  <div class="flex flex-col gap-y-4 mt-6" style:width="400px">
    {#if step === AuthStep.Base}
      {#each LOGIN_OPTIONS as { label, icon, style, connection } (connection)}
        <CtaButton
          variant={style === "primary" ? "primary" : "secondary"}
          on:click={() => {
            webAuth.authorize({ connection });
          }}
        >
          <div class="flex justify-center items-center gap-x-2 font-medium">
            {#if icon}
              <svelte:component this={icon} />
            {/if}
            <div>{label}</div>
          </div>
        </CtaButton>
      {/each}

      <OrSeparator />

      <EmailSubmissionForm on:submit={processEmailSubmission} />
    {/if}

    {#if step === AuthStep.Login || step === AuthStep.SignUp}
      <EmailPasswordForm
        {email}
        isDomainDisabled={domainDisabled}
        {webAuth}
        on:back={backToBaseStep}
      />
    {/if}
  </div>

  {#if errorText}
    <div style:max-width="400px" class="text-red-500 text-sm mt-3">
      {errorText}
    </div>
  {/if}

  <Disclaimer />
  <div class="mt-6 text-center">
    <CtaNeedHelp />
  </div>
</AuthContainer>
