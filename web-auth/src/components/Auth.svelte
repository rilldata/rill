<script lang="ts">
  import auth0, { WebAuth } from "auth0-js";
  import { onMount } from "svelte";
  import { LOGIN_OPTIONS } from "../config";
  import RillLogoSquareNegative from "@rilldata/web-common/components/icons/RillLogoSquareNegative.svelte";
  import CtaButton from "@rilldata/web-common/components/calls-to-action/CTAButton.svelte";
  import AuthContainer from "./AuthContainer.svelte";
  import Disclaimer from "./Disclaimer.svelte";
  import EmailPassForm from "./EmailPassForm.svelte";
  import RillTheme from "@rilldata/web-common/layout/RillTheme.svelte";

  export let configParams: string;
  export let cloudClientIDs = "";
  export let oktaName = "";
  export let pingFedName = "";
  export let disableForgotPassDomains = "";

  const cloudClientIDsArr = cloudClientIDs.split(",");
  const disableForgotPassDomainsArr = disableForgotPassDomains.split(",");

  // By default show the LogIn page
  let isLoginPage = true;
  let errorText = "";

  let webAuth: WebAuth;
  const databaseConnection = "Username-Password-Authentication";

  $: loginOptions = LOGIN_OPTIONS;

  $: loginOptions.forEach((option) => {
    if (option.name === "Okta") {
      option.connection = oktaName;
    }
    if (option.name === "Pingfed") {
      option.connection = pingFedName;
    }
  });

  function initConfig() {
    const config = JSON.parse(
      decodeURIComponent(escape(window.atob(configParams)))
    );

    if (cloudClientIDsArr.includes(config?.clientID)) {
      loginOptions = loginOptions.filter(
        (option) => !["Okta", "Pingfed"].includes(option.name)
      );
    }

    const params = Object.assign(
      {
        overrides: {
          __tenant: config.auth0Tenant,
          __token_issuer: "auth.rilldata.io",
        },
        domain: config.auth0Domain,
        clientID: config.clientID,
        redirectUri: config.callbackURL,
        responseType: "code",
      },
      config.internalOptions
    );

    webAuth = new auth0.WebAuth(params);
  }

  function displayError(err: any) {
    errorText = err.message;
  }

  function authorize(connection: string) {
    webAuth.authorize({ connection });
  }

  function handleEmailSubmit(email: string, password: string) {
    errorText = "";
    if (isLoginPage) {
      webAuth.login(
        {
          realm: databaseConnection,
          username: email,
          password: password,
        },
        (err) => {
          if (err) displayError(err);
        }
      );
    } else {
      webAuth.redirect.signupAndLogin(
        {
          connection: databaseConnection,
          email: email,
          password: password,
        },
        (err) => {
          if (err) displayError(err);
        }
      );
    }
  }

  function handleResetPassword(email: string) {
    errorText = "";
    if (!email) return displayError({ message: "Please enter an email" });

    if (
      disableForgotPassDomainsArr.some((domain) =>
        email.toLowerCase().endsWith(domain.toLowerCase())
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
        if (err) displayError(err);
        else alert(resp);
      }
    );
  }

  onMount(() => {
    initConfig();
  });
</script>

<RillTheme>
  <AuthContainer>
    <RillLogoSquareNegative size="84px" />
    <div class="text-xl my-6">
      {isLoginPage ? "Log in to Rill" : "Create your Rill account"}
    </div>
    <div class="flex flex-col gap-y-4" style:width="400px">
      {#each loginOptions as { label, icon, style, connection }}
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
      <EmailPassForm
        {isLoginPage}
        on:submit={(e) => {
          handleEmailSubmit(e.detail.email, e.detail.password);
        }}
        on:resetPass={(e) => {
          handleResetPassword(e.detail.email);
        }}
      />
    </div>

    {#if errorText}
      <div class="text-red-500 text-sm mt-2">{errorText}</div>
    {/if}

    <Disclaimer />
    <div class="mt-6 text-sm text-slate-500">
      {isLoginPage ? "Don't" : "Already"} have an account?

      <!-- svelte-ignore a11y-invalid-attribute -->
      <a href="#" on:click={() => (isLoginPage = !isLoginPage)}>
        {isLoginPage ? "Sign up" : "Log in"}</a
      >
    </div>
  </AuthContainer>
</RillTheme>
