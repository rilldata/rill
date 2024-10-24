<script lang="ts">
  import CtaButton from "@rilldata/web-common/components/calls-to-action/CTAButton.svelte";
  import RillLogoSquareNegative from "@rilldata/web-common/components/icons/RillLogoSquareNegative.svelte";
  import auth0, { WebAuth } from "auth0-js";
  import type { AuthOptions } from "auth0-js";
  import { onMount } from "svelte";
  import { LOGIN_OPTIONS } from "../config";
  import AuthContainer from "./AuthContainer.svelte";
  import EmailPassForm from "./EmailPassForm.svelte";
  import { getConnectionFromEmail } from "./utils";
  import OrSeparator from "./OrSeparator.svelte";
  import SSOForm from "./SSOForm.svelte";
  import EmailSubmissionForm from "./EmailSubmissionForm.svelte";
  import DiscordCTA from "./DiscordCTA.svelte";
  import Disclaimer from "./Disclaimer.svelte";
  import Spacer from "./Spacer.svelte";

  type InternalOptions = {
    protocol: string;
    response_type: string;
    prompt: string;
    scope: string;
    _csrf: string;
    leeway: number;
  };

  type Config = {
    auth0Domain: string;
    clientID: string;
    auth0Tenant: string;
    authorizationServer: {
      issuer: string;
    };
    callbackURL: string;
    internalOptions: InternalOptions;
    extraParams?: { screen_hint?: string };
  };

  export let configParams: string;
  export let disableForgotPassDomains = "";
  export let connectionMap = "{}";

  const LOCAL_STORAGE_KEY = "last_used_connection";
  const DATABASE_CONNECTION = "Username-Password-Authentication";

  const connectionMapObj = JSON.parse(connectionMap);
  const disableForgotPassDomainsArr = disableForgotPassDomains.split(",");

  enum AuthStep {
    Base = 0,
    SSO = 1,
    EmailPassword = 2,
    SignUp = 3,
    Thanks = 4,
  }

  $: errorText = "";

  let isSSODisabled = false;
  let isEmailDisabled = false;
  let lastUsedConnection: string | null = null;
  let email = "";
  let step: AuthStep = AuthStep.Base;
  let webAuth: WebAuth;

  function getLastUsedConnection() {
    return localStorage.getItem(LOCAL_STORAGE_KEY);
  }

  function setLastUsedConnection(connection: string | null) {
    if (connection) {
      localStorage.setItem(LOCAL_STORAGE_KEY, connection);
    } else {
      localStorage.removeItem(LOCAL_STORAGE_KEY);
    }
    lastUsedConnection = connection;
  }

  $: {
    const storedConnection = getLastUsedConnection();
    if (storedConnection) {
      lastUsedConnection = storedConnection;
    } else {
      setLastUsedConnection(null);
    }
  }

  function initConfig() {
    const config = JSON.parse(
      decodeURIComponent(escape(window.atob(configParams))),
    ) as Config;

    if (config?.extraParams?.screen_hint === "signup") {
      step = AuthStep.SignUp;
    }

    const params: AuthOptions = Object.assign(
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

    webAuth = new auth0.WebAuth(params);
  }

  function displayError(err: any) {
    errorText = err.message;
  }

  function authorize(connection: string) {
    setLastUsedConnection(connection);
    webAuth.authorize({ connection });
  }

  function handleSSOLogin(email: string) {
    isSSODisabled = true;
    errorText = "";

    const connectionName = getConnectionFromEmail(email, connectionMapObj);

    if (!connectionName) {
      displayError({
        message: `IDP for the email ${email} not found. Please contact your administrator.`,
      });
      isSSODisabled = false;
      return;
    }

    webAuth.authorize({
      connection: connectionName,
      login_hint: email,
      prompt: "login",
    });

    setLastUsedConnection(connectionName);
  }

  function handleEmailSubmit(email: string, password: string) {
    isEmailDisabled = true;
    errorText = "";

    function handleAuthError(err: any) {
      // Auth0 is not consistent in the naming of the error description field
      const errorText =
        typeof err?.description === "string"
          ? err.description
          : typeof err?.policy === "string"
            ? err.policy
            : typeof err?.error_description === "string"
              ? err.error_description
              : err?.message;

      displayError({ message: errorText });
      isEmailDisabled = false;
    }

    try {
      webAuth.login(
        {
          realm: DATABASE_CONNECTION,
          username: email,
          password: password,
        },
        (err) => {
          if (err) {
            console.log("login err", err);
            // TODO: revisit error message from staging
            // Check if the error indicates the user does not exist
            if (err.error === "user_not_found") {
              // Attempt to sign up the user
              webAuth.redirect.signupAndLogin(
                {
                  connection: DATABASE_CONNECTION,
                  email: email,
                  password: password,
                },
                (signupErr: any) => {
                  if (signupErr) handleAuthError(signupErr);
                  else isEmailDisabled = false;
                },
              );
            } else {
              handleAuthError(err);
            }
          } else {
            isEmailDisabled = false;
          }
        },
      );
    } catch (err) {
      handleAuthError(err);
    }
  }

  function handleResetPassword(email: string) {
    errorText = "";
    if (!email) return displayError({ message: "Please enter an email" });

    if (
      disableForgotPassDomainsArr.some((domain) =>
        email.toLowerCase().endsWith(domain.toLowerCase()),
      )
    ) {
      return displayError({
        message: "Password reset is not available. Please contact your admin.",
      });
    }

    webAuth.changePassword(
      {
        connection: DATABASE_CONNECTION,
        email: email,
      },
      (err, resp) => {
        if (err) displayError({ message: err?.description });
        else alert(resp);
      },
    );
  }

  function handleEmailSubmission(event) {
    email = event.detail.email;
    const connectionName = getConnectionFromEmail(email, connectionMapObj);

    if (connectionName) {
      step = AuthStep.SSO;
    } else {
      step = AuthStep.EmailPassword;
    }
  }

  function getHeadingText(step: AuthStep): string {
    switch (step) {
      case AuthStep.Base:
        return "Log in or sign up";
      case AuthStep.SSO:
        return "Log in with SSO";
      case AuthStep.EmailPassword:
        return "Log in with email";
      case AuthStep.SignUp:
        return "Sign up with email";
      case AuthStep.Thanks:
        return "Thanks for signing up!";
      default:
        return "";
    }
  }

  function getSubheadingText(step: AuthStep, email: string): string {
    switch (step) {
      case AuthStep.SSO:
        return `SAML SSO enabled workspace is associated with ${email}`;
      case AuthStep.EmailPassword:
        return `Log in using ${email}`;
      default:
        return "";
    }
  }

  $: headingText = getHeadingText(step);
  $: subheadingText = getSubheadingText(step, email);

  onMount(() => {
    initConfig();
  });
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
        {subheadingText}
      </div>
    {:else}
      <Spacer />
    {/if}
  </div>

  <div class="flex flex-col gap-y-4 mt-6" style:width="400px">
    {#if lastUsedConnection}
      <div class="text-sm text-gray-500">
        Last used connection: {lastUsedConnection}
      </div>
    {/if}
    {#if step === AuthStep.Base}
      {#each LOGIN_OPTIONS as { label, icon, style, connection } (connection)}
        <CtaButton
          variant={style === "primary" ? "primary" : "secondary"}
          on:click={() => authorize(connection)}
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

      <EmailSubmissionForm
        disabled={isEmailDisabled}
        on:submitEmail={handleEmailSubmission}
      />
    {/if}

    {#if step === AuthStep.SSO}
      <SSOForm
        disabled={isSSODisabled}
        on:submitSSO={(e) => {
          handleSSOLogin(e.detail);
        }}
        on:back={() => {
          step = AuthStep.Base;
        }}
      />
    {/if}

    <!-- TODO: only show forget password in sign up flow -->
    {#if step === AuthStep.EmailPassword}
      <EmailPassForm
        disabled={isEmailDisabled}
        {email}
        on:submit={(e) => {
          handleEmailSubmit(e.detail.email, e.detail.password);
        }}
        on:resetPass={(e) => {
          handleResetPassword(e.detail.email);
        }}
        on:back={() => {
          step = AuthStep.Base;
        }}
        showForgetPassword={step === AuthStep.EmailPassword}
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
    <DiscordCTA />
  </div>
</AuthContainer>
