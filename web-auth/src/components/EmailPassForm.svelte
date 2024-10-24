<script lang="ts">
  import CtaButton from "@rilldata/web-common/components/calls-to-action/CTAButton.svelte";
  import Eye from "@rilldata/web-common/components/icons/Eye.svelte";
  import EyeInvisible from "@rilldata/web-common/components/icons/EyeInvisible.svelte";
  import { createEventDispatcher } from "svelte";
  import { validateEmail } from "./utils";
  import { ArrowLeftIcon } from "lucide-svelte";

  const dispatch = createEventDispatcher();

  export let disabled = false;
  export let email = "";
  export let showForgetPassword = false;

  let password = "";
  let showForm = false;
  let showPassword = false;
  let errorText = "";

  let inputClasses =
    "h-10 px-4 py-2 border border-slate-300 rounded-sm text-base";
  let focusClasses =
    "ring-offset-2 focus:ring-2 focus:ring-primary-ry-300 focus:outline-none";

  $: type = showPassword ? "text" : "password";

  function onPassInput(e: any) {
    if (e.key === "Enter") {
      handleSubmit();
    }
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

  $: disabled = !(password.length > 0);
</script>

<form on:submit={handleSubmit} class="flex flex-col gap-y-4">
  <div class="relative">
    <!-- TODO: look into using <Input /> component -->
    <input
      class="{inputClasses} {focusClasses}"
      style:width="400px"
      {type}
      on:input={onPassInput}
      id="password"
      placeholder="Password"
    />

    {#if errorText}
      <div class="text-red-500 text-sm -mt-2">
        {errorText}
      </div>
    {/if}

    <!-- svelte-ignore a11y-click-events-have-key-events -->
    <span
      role="button"
      tabindex="0"
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
  {#if showForgetPassword}
    <div>
      <button
        type="button"
        on:click={handleForgotPass}
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
