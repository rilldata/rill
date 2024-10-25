<script lang="ts">
  import CtaButton from "@rilldata/web-common/components/calls-to-action/CTAButton.svelte";
  import Eye from "@rilldata/web-common/components/icons/Eye.svelte";
  import EyeInvisible from "@rilldata/web-common/components/icons/EyeInvisible.svelte";
  import { createEventDispatcher, onMount } from "svelte";
  import { ArrowLeftIcon } from "lucide-svelte";
  import { WebAuth } from "auth0-js";
  import { DATABASE_CONNECTION } from "../constants";
  import { AuthStep } from "../types";
  import type { Auth0Error } from "auth0-js";

  const dispatch = createEventDispatcher();

  export let disabled = false;
  export let email = "";
  export let showForgetPassword = false;
  export let isDomainDisabled = false;
  export let isEmailDisabled = false;
  export let webAuth: WebAuth;
  export let step: AuthStep;

  let password = "";
  let showPassword = false;
  let errorText = "";

  let inputClasses =
    "h-10 px-4 py-2 border border-slate-300 rounded-sm text-base";
  let focusClasses =
    "ring-offset-2 focus:ring-2 focus:ring-primary-ry-300 focus:outline-none";

  $: type = showPassword ? "text" : "password";

  function onPassInput(e: any) {
    password = e.target.value;
  }

  function handleSubmit() {
    if (!password) {
      errorText = "Please enter your password";
      return;
    }

    errorText = "";

    authenticateUser(email, password);
  }

  function displayError(err: any) {
    errorText = err.message;
  }

  function handleResetPassword() {
    errorText = "";

    if (!email) return displayError({ message: "Please enter an email" });
    if (isDomainDisabled) {
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

  function handleAuthError(err: Auth0Error) {
    // NOTE: Auth0 is not consistent in the naming of the error description field
    const errorText =
      typeof err?.description === "string"
        ? err.description
        : typeof err?.policy === "string"
          ? err.policy
          : typeof err?.error_description === "string"
            ? err.error_description
            : err?.errorDescription;

    displayError({ message: errorText });
    isEmailDisabled = false;
  }

  function authenticateUser(email: string, password: string) {
    isEmailDisabled = true;
    errorText = "";

    try {
      if (step === AuthStep.SignUp) {
        // Directly attempt to sign up and log in the user
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
        webAuth.login(
          {
            realm: DATABASE_CONNECTION,
            username: email,
            password: password,
          },
          (err) => {
            if (err) displayError({ message: err?.description });
            isEmailDisabled = false;
          },
        );
      }
    } catch (err) {
      handleAuthError(err);
    }
  }

  $: disabled = !(password.length > 0);
</script>

<form on:submit|preventDefault={handleSubmit} class="flex flex-col gap-y-4">
  <div>
    <div class="relative flex items-center" style="max-width: 400px;">
      <input
        class="{inputClasses} {focusClasses} flex-grow pr-10"
        style:width="100%"
        {type}
        on:input={onPassInput}
        id="password"
        placeholder="Password"
      />

      <!-- svelte-ignore a11y-click-events-have-key-events -->
      <span
        role="button"
        tabindex="0"
        class="absolute right-3 cursor-pointer"
        on:click={() => (showPassword = !showPassword)}
      >
        {#if !showPassword}
          <Eye />
        {:else}
          <EyeInvisible />
        {/if}
      </span>
    </div>

    {#if errorText}
      <div style:max-width="400px" class="text-red-500 text-sm mt-3">
        {errorText}
      </div>
    {/if}
  </div>

  {#if showForgetPassword}
    <div class="text-left">
      <button
        type="button"
        on:click={handleResetPassword}
        class="text-sm text-slate-500 pl-1 font-medium">Forgot password?</button
      >
    </div>
  {/if}

  <CtaButton {disabled} variant="primary" submitForm>
    <div class="flex justify-center font-medium">Continue</div>
  </CtaButton>
  <CtaButton
    variant="secondary"
    gray
    on:click={() => {
      dispatch("back");
    }}
  >
    <div class="flex justify-center items-center font-medium">
      <ArrowLeftIcon class="mr-1" size={14} />
      <span>Back</span>
    </div>
  </CtaButton>
</form>
