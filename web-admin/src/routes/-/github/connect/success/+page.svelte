<script>
  import Github from "@rilldata/web-common/components/icons/Github.svelte";
  import CtaButton from "../../../../../components/CTAButton.svelte";

  // https://stackoverflow.com/questions/8057802/html-button-close-window
  function closePage() {
    window.open("", "_self", "");
    window.close();
  }
</script>

<svelte:head>
  <title>Github connected successfully</title>
</svelte:head>

<div class="flex flex-col justify-center items-center h-3/5">
  <div class="flex flex-col justify-center items-center gap-y-6 w-[400px]">
    <Github className="w-[56px] h-[42px] text-gray-900" />
    <h1 class="text-xl font-medium text-gray-800">
      Github connected successfully
    </h1>
    <p class="text-base text-gray-500">
      Close this page and continue setup in the Rill CLI.
    </p>
    <CtaButton variant="secondary" on:click={closePage}>Close page</CtaButton>
  </div>
</div>
