<script>
  import Github from "@rilldata/web-common/components/icons/Github.svelte";
  import CtaButton from "../../../../../components/calls-to-action/CTAButton.svelte";
  import CtaContentContainer from "../../../../../components/calls-to-action/CTAContentContainer.svelte";
  import CtaLayoutContainer from "../../../../../components/calls-to-action/CTALayoutContainer.svelte";

  // https://stackoverflow.com/questions/8057802/html-button-close-window
  function closePage() {
    window.open("", "_self", "");
    window.close();
  }
</script>

<svelte:head>
  <title>Github connected successfully</title>
</svelte:head>

<CtaLayoutContainer>
  <CtaContentContainer>
    <Github className="w-10 h-10 text-gray-900" />
    <h1 class="text-xl font-medium text-gray-800">
      Github connected successfully
    </h1>
    <p class="text-base text-gray-500">
      Close this page and continue setup in the Rill CLI.
    </p>
    <CtaButton variant="secondary" on:click={closePage}>Close page</CtaButton>
  </CtaContentContainer>
</CtaLayoutContainer>
