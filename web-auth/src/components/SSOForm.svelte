<script lang="ts">
  import { slide } from "svelte/transition";
  import { createEventDispatcher } from "svelte";
  import CtaButton from "@rilldata/web-common/components/calls-to-action/CTAButton.svelte";
  import { validateEmail } from "./utils";

  const dispatch = createEventDispatcher();

  export let disabled = false;

  let userEmail = "";
  let showForm = false;
  let errorText = "";

  let inputClasses =
    "h-10 px-4 py-2 border border-slate-300 rounded-sm text-base";
  let focusClasses =
    "ring-offset-2 focus:ring-2 focus:ring-blue-300 focus:outline-none";

  function handleSubmit() {
    if (!showForm) {
      showForm = true;
      return;
    }

    if (!userEmail) {
      return;
    }

    if (!validateEmail(userEmail)) {
      errorText = "Please enter a valid email address";
      return;
    }

    errorText = "";

    dispatch("ssoSubmit", userEmail.toLowerCase());
  }

  function handleKeydown(e) {
    if (e.key === "Enter") {
      handleSubmit();
    }
  }
</script>

<!-- "mb-6" -->
<div class:mb-6={showForm}>
  {#if showForm}
    <div class="mt-6 mb-4 flex flex-col gap-y-4" transition:slide>
      <input
        class="{inputClasses} {focusClasses}"
        style:width="400px"
        type="email"
        placeholder="Enter your email address"
        id="sso"
        bind:value={userEmail}
        on:keydown={handleKeydown}
      />
    </div>
  {/if}

  <CtaButton {disabled} variant="secondary" on:click={() => handleSubmit()}>
    <div class="flex justify-center font-medium w-[400px]">
      <div>Continue with SAML SSO</div>
    </div>
  </CtaButton>
  {#if errorText}
    <div class="mt-2 text-red-500 text-sm">{errorText}</div>
  {/if}
</div>
