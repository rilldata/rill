<script lang="ts">
  import CtaButton from "@rilldata/web-common/components/calls-to-action/CTAButton.svelte";
  import { validateEmail } from "./utils";
  import { createEventDispatcher } from "svelte";

  export let disabled = false;

  let email = "";
  let haveValidEmail = false;
  let errorText = "";

  const dispatch = createEventDispatcher();

  let inputClasses =
    "h-10 px-4 py-2 border border-slate-300 rounded-sm text-base";
  let focusClasses =
    "ring-offset-2 focus:ring-2 focus:ring-primary-ry-300 focus:outline-none";

  function handleContinue() {
    if (!email) {
      errorText = "Please enter your email";
      return;
    }

    if (!validateEmail(email)) {
      haveValidEmail = false;
      errorText = "Please enter a valid email address";
      return;
    }

    errorText = "";

    haveValidEmail = true;
    dispatch("emailSubmit", { email });
  }

  $: {
    if (validateEmail(email)) {
      disabled = false;
    } else {
      disabled = true;
    }
  }
</script>

<div>
  <div class="mb-4 flex flex-col gap-y-4">
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
  </div>

  <CtaButton
    {disabled}
    variant="secondary"
    on:click={() => {
      if (email) {
        handleContinue();
      }
    }}
  >
    <div class="flex justify-center font-medium">
      <div>Continue</div>
    </div>
  </CtaButton>
</div>
