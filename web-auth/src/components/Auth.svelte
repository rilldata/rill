<script lang="ts">
  import CtaButton from "@rilldata/web-common/components/calls-to-action/CTAButton.svelte";
  import RillLogoSquareNegative from "@rilldata/web-common/components/icons/RillLogoSquareNegative.svelte";
  import auth0, { WebAuth } from "auth0-js";
  import type { AuthOptions } from "auth0-js";
  import { onMount } from "svelte";
  import { LOGIN_OPTIONS } from "../config";
  import AuthContainer from "./AuthContainer.svelte";
  import { getConnectionFromEmail } from "./utils";
  import OrSeparator from "./OrSeparator.svelte";
  import EmailSubmissionForm from "./EmailSubmissionForm.svelte";
  import Disclaimer from "./Disclaimer.svelte";
  import Spacer from "./Spacer.svelte";
  import { AuthStep, type Config } from "../types";
  import CtaNeedHelp from "@rilldata/web-common/components/calls-to-action/CTANeedHelp.svelte";

  export let configParams: string;
  export let cloudClientIDs = "";
  export let connectionMap = "{}";

  const connectionMapObj = JSON.parse(connectionMap);
  const cloudClientIDsArr = cloudClientIDs.split(",");

  $: errorText = "";

  let email = "";
  let step: AuthStep = AuthStep.Base;
  let webAuth: WebAuth;

  $: isSignup = false;
  $: isRillDash = false;

  let verifying = false;
  let verificationCode: string = "";

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

      isSignup = config?.extraParams?.screen_hint === "signup";

      if (cloudClientIDsArr.includes(config?.clientID)) {
        isRillDash = true;
      }

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
      errorText = "Failed to initialize authentication";
    }
  }

  function authorizeSSO(email: string, connectionName: string) {
    // See: https://community.auth0.com/t/home-realm-discovery-using-auth0-js/17643/2
    webAuth.authorize({
      connection: connectionName,
      login_hint: email,
      prompt: "login",
    });
  }

  function startPasswordless(email: string) {
    webAuth.passwordlessStart(
      {
        connection: "email",
        send: "code",
        email: email,
      },
      (err) => {
        if (err) {
          errorText = err.description || "An error occurred";
          return;
        }
        step = AuthStep.Thanks;
      },
    );
  }

  function processEmailSubmission(event) {
    email = event.detail.email;

    const connectionName = getConnectionFromEmail(email, connectionMapObj);

    if (connectionName) {
      authorizeSSO(email, connectionName);
    } else {
      startPasswordless(email);
    }
  }

  function getHeadingText(step: AuthStep): string {
    switch (step) {
      case AuthStep.Base:
        return "Log in or sign up";
      case AuthStep.Thanks:
        return "Check your email";
      default:
        return "";
    }
  }

  function getSubheadingText(step: AuthStep, email: string): string {
    switch (step) {
      case AuthStep.Thanks:
        return `We sent a verification code to <span class="font-medium">${email}</span>`;
      default:
        return "";
    }
  }

  function backToBaseStep() {
    step = AuthStep.Base;
  }

  function parseQueryString() {
    const params = new URLSearchParams(window.location.search);
    return {
      code: params.get("code"),
      email: params.get("email"),
    };
  }

  function verifyPasswordlessCode(email: string, code: string) {
    verifying = true;
    errorText = "";

    webAuth.passwordlessVerify(
      {
        connection: "email",
        email: email,
        verificationCode: code,
      },
      (err) => {
        verifying = false;
        if (err) {
          errorText = err.description || "Failed to verify email code";
          console.error("Verification error:", err);
        }
      },
    );
  }

  function handleCodeSubmit(code: string) {
    verifyPasswordlessCode(email, code);
  }

  onMount(() => {
    initConfig();

    const { code, email } = parseQueryString();
    if (code && email) {
      verifyPasswordlessCode(email, code);
    }
  });

  $: headingText = getHeadingText(step);
  $: subheadingText = getSubheadingText(step, email);
</script>

<AuthContainer>
  <RillLogoSquareNegative size="84px" />
  <Spacer />
  <div class="flex flex-col items-center gap-y-2 text-center">
    <div class="text-xl text-slate-800">
      {headingText}
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
    {#if verifying}
      <div class="text-center text-gray-600">Verifying your email...</div>
    {:else if step === AuthStep.Base}
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

    {#if step === AuthStep.Thanks}
      <div class="text-center text-gray-600">
        Enter the verification code sent to your email
      </div>
      <div class="flex flex-col gap-y-4">
        <input
          type="text"
          bind:value={verificationCode}
          placeholder="Enter verification code"
          class="p-2 border rounded"
        />
        <CtaButton
          variant="primary"
          on:click={() => handleCodeSubmit(verificationCode)}
        >
          Verify Code
        </CtaButton>
        <CtaButton variant="secondary" on:click={backToBaseStep}>
          Use a different email
        </CtaButton>
      </div>
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
