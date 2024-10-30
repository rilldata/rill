<script lang="ts">
  import { page } from "$app/stores";
  import GithubSuccess from "@rilldata/web-common/components/icons/GithubSuccess.svelte";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaMessage from "@rilldata/web-common/components/calls-to-action/CTAMessage.svelte";
  import { onMount } from "svelte";

  onMount(() => {
    if ($page.url.searchParams.get("autoclose") === "true") {
      // close the window autoclose was set.
      // this happens when the connection was requested through UI flow
      window.close();
    }
  });
</script>

<svelte:head>
  <title>GitHub connected successfully</title>
</svelte:head>

<CtaLayoutContainer>
  <CtaContentContainer>
    <GithubSuccess />
    <CtaHeader>GitHub connected successfully</CtaHeader>
    <CtaMessage>Close this page and continue setup in the Rill CLI.</CtaMessage>
  </CtaContentContainer>
</CtaLayoutContainer>
