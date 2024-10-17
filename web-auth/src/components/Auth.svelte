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
  import EmailSubmission from "./EmailSubmission.svelte";
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
  export let cloudClientIDs = "";
  export let disableForgotPassDomains = "";
  export let connectionMap = "{}";

  const LOCAL_STORAGE_KEY = "last_used_connection";

  const connectionMapObj = JSON.parse(connectionMap);
  const cloudClientIDsArr = cloudClientIDs.split(",");
  const disableForgotPassDomainsArr = disableForgotPassDomains.split(",");

  const databaseConnection = "Username-Password-Authentication";

  enum AuthStep {
    Base = 0,
    SSO = 1,
    EmailPassword = 2,
    SignUp = 3,
    Thanks = 4,
  }

  $: errorText = "";
  $: isRillCloud = false;

  let isSSODisabled = false;
  let isEmailDisabled = false;
  let lastUsedConnection: string | null = null;

  let email = "";
  let emailSubmitted = false;

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
    if (cloudClientIDsArr.includes(config?.clientID)) {
      isRillCloud = true;
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
    try {
      webAuth.login(
        {
          realm: databaseConnection,
          username: email,
          password: password,
        },
        (err) => {
          if (err) displayError({ message: err?.description });
          isEmailDisabled = false;
        },
      );

      setLastUsedConnection(databaseConnection);

      // TO BE REMOVED
      // TODO: should we check for `last_used_connection`
      // webAuth.redirect.signupAndLogin(
      //   {
      //     connection: databaseConnection,
      //     email: email,
      //     password: password,
      //   },
      //   // explicitly typing as any to avoid missing property TS/svelte-check error
      //   (err: any) => {
      //     // Auth0 is not consistent in the naming of the error description field
      //     const errorText =
      //       typeof err?.description === "string"
      //         ? err.description
      //         : typeof err?.policy === "string"
      //           ? err.policy
      //           : typeof err?.error_description === "string"
      //             ? err.error_description
      //             : err?.message;

      //     if (err) displayError({ message: errorText });
      //     isEmailDisabled = false;
      //   },
      // );
    } catch (err) {
      displayError({ message: err?.description || err?.policy });
      isEmailDisabled = false;
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
        connection: databaseConnection,
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
    emailSubmitted = true;

    if (connectionName) {
      step = AuthStep.SSO;
    } else {
      step = AuthStep.EmailPassword;
    }
  }

  function getHeadingText() {
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
    }
  }

  function getSubheadingText() {
    switch (step) {
      case AuthStep.SSO:
        return `SAML SSO enabled workspace is associated with ${email}`;
      case AuthStep.EmailPassword:
        return `Log in using ${email}`;
    }
  }

  $: headingText = getHeadingText();
  $: subheadingText = getSubheadingText();

  $: console.log("headingText: ", headingText);

  onMount(() => {
    initConfig();
  });
</script>

<RillTheme>
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

        <EmailSubmission
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
        />
      {/if}

      <!-- TODO: AuthStep.SignUp -->
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
</RillTheme>
