<script lang="ts">
  import { slide } from "svelte/transition";
  import { createEventDispatcher } from "svelte";
  import EyeInvisible from "@rilldata/web-common/components/icons/EyeInvisible.svelte";
  import Eye from "@rilldata/web-common/components/icons/Eye.svelte";
  import CtaButton from "@rilldata/web-common/components/calls-to-action/CTAButton.svelte";
  import { validateEmail } from "./utils";

  const dispatch = createEventDispatcher();

  export let isLoginPage = false;
  export let disabled = false;

  let email = "";
  let password = "";

  let showForm = false;
  let showPassword = false;
  let errorText = "";

  let inputClasses =
    "h-10 px-4 py-2 border border-slate-300 rounded-sm text-base";
  let focusClasses =
    "ring-offset-2 focus:ring-2 focus:ring-blue-300 focus:outline-none";

  $: type = showPassword ? "text" : "password";

  function onPassInput(e: any) {
    password = e.target.value;
  }

  function handleSubmit() {
    if (!showForm) {
      showForm = true;
      return;
    }

    if (!email || !password) {
      errorText = "Please enter your email and password";
      return;
    }

    if (!validateEmail(email)) {
      errorText = "Please enter a valid email address";
      return;
    }

    errorText = "";

    dispatch("submit", {
      email,
      password,
    });
  }

  function handleForgotPass() {
    if (!validateEmail(email)) {
      errorText = "Please enter a valid email address";
      return;
    }

    errorText = "";
    dispatch("resetPass", {
      email,
    });
  }
</script>

<div>
  {#if showForm}
    <div class="mt-6 mb-4 flex flex-col gap-y-4" transition:slide>
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

      <div style="position: relative;">
        <input
          class="{inputClasses} {focusClasses}"
          style:width="400px"
          {type}
          on:input={onPassInput}
          id="password"
          placeholder={isLoginPage ? "Enter your password" : "Create password"}
        />

        <!-- svelte-ignore a11y-click-events-have-key-events -->
        <span
          style:right="10px"
          class="absolute top-1/2 transform -translate-y-1/2 cursor-pointer"
          on:click={() => (showPassword = !showPassword)}
        >
          {#if !showPassword}
            <Eye />
          {:else}
            <EyeInvisible />
          {/if}
        </span>
      </div>
    </div>
    {#if isLoginPage}
      <div>
        <button
          on:click={() => handleForgotPass()}
          class="text-sm mb-5 text-slate-500 pl-1 font-medium"
          >Forgot password?</button
        >
      </div>
    {/if}
  {/if}

  <CtaButton {disabled} variant="secondary" on:click={() => handleSubmit()}>
    <div class="flex justify-center font-medium w-[400px]">
      <div>Continue with Email</div>
    </div>
  </CtaButton>
</div>
