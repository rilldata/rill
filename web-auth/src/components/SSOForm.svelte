<script lang="ts">
  import { slide } from "svelte/transition";
  import { createEventDispatcher } from "svelte";
  import CtaButton from "@rilldata/web-common/components/calls-to-action/CTAButton.svelte";

  const dispatch = createEventDispatcher();

  export let disabled = false;

  let companySlug = "";
  let showForm = false;

  let inputClasses =
    "h-10 px-4 py-2 border border-slate-300 rounded-sm text-base";
  let focusClasses =
    "ring-offset-2 focus:ring-2 focus:ring-blue-300 focus:outline-none";

  function handleSubmit() {
    if (!showForm) {
      showForm = true;
      return;
    }

    if (!companySlug) {
      return;
    }

    dispatch("ssoSubmit", companySlug.toLowerCase());
  }
</script>

<div class:mb-6={showForm}>
  {#if showForm}
    <div class="mt-6 mb-4 flex flex-col gap-y-4" transition:slide>
      <input
        class="{inputClasses} {focusClasses}"
        style:width="400px"
        type="text"
        placeholder="Enter your organization slug"
        id="sso"
        bind:value={companySlug}
      />
    </div>
  {/if}

  <CtaButton {disabled} variant="secondary" on:click={() => handleSubmit()}>
    <div class="flex justify-center font-medium w-[400px]">
      <div>Continue with SAML SSO</div>
    </div>
  </CtaButton>
</div>
