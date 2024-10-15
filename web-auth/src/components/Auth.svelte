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

  const connectionMapObj = JSON.parse(connectionMap) as Record<
    string,
    string[]
  >;
  const cloudClientIDsArr = cloudClientIDs.split(",");
  const disableForgotPassDomainsArr = disableForgotPassDomains.split(",");

  $: errorText = "";
  $: isRillCloud = false;

  let isSSODisabled = false;
  let isEmailDisabled = false;

  let webAuth: WebAuth;
  const databaseConnection = "Username-Password-Authentication";

  function initConfig() {
    const config = JSON.parse(
      decodeURIComponent(escape(window.atob(configParams))),
    ) as Config;

    // TO BE REMOVED
    // if (config?.extraParams?.screen_hint === "signup") {
    //   isLoginPage = false;
    // }

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

  onMount(() => {
    initConfig();
  });
</script>

<RillTheme>
  <AuthContainer>
    <RillLogoSquareNegative size="84px" />
    <div class="text-xl my-6">Log in or sign up</div>
    <div class="flex flex-col gap-y-4 mt-6" style:width="400px">
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

      <div class="flex items-center justify-center h-6">
        <hr class="flex-grow border-t border-slate-300" />
        <span class="px-2 text-slate-500 text-sm">or</span>
        <hr class="flex-grow border-t border-slate-300" />
      </div>

      <!-- TODO: re-enable SSO -->
      <!-- <SSOForm
        disabled={isSSODisabled}
        on:ssoSubmit={(e) => {
          handleSSOLogin(e.detail);
        }}
      /> -->

      <EmailPassForm
        disabled={isEmailDisabled}
        on:submit={(e) => {
          handleEmailSubmit(e.detail.email, e.detail.password);
        }}
        on:resetPass={(e) => {
          handleResetPassword(e.detail.email);
        }}
      />
    </div>

    {#if errorText}
      <div style:max-width="400px" class="text-red-500 text-sm mt-3">
        {errorText}
      </div>
    {/if}

    <!-- REVISIT AFTER https://www.figma.com/design/Qt6EyotCBS3V6O31jVhMQ7?node-id=18329-561704#987505195 -->
    <!-- <Disclaimer /> -->

    <div class="mt-6 text-center">
      <p class="text-sm text-gray-500">
        Need help? Reach out to us on <a
          href="http://bit.ly/3jg4IsF"
          target="_blank">Discord</a
        >
      </p>
    </div>
  </AuthContainer>
</RillTheme>
