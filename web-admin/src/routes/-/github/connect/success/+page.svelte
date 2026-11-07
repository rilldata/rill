<script lang="ts">
  import { page } from "$app/stores";
  import GithubSuccess from "@rilldata/web-common/components/icons/GithubSuccess.svelte";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaMessage from "@rilldata/web-common/components/calls-to-action/CTAMessage.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
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
  <title>{m.github_connected_successfully()}</title>
</svelte:head>

<CtaLayoutContainer>
  <CtaContentContainer>
    <GithubSuccess />
    <CtaHeader>{m.github_connected_successfully()}</CtaHeader>
    <CtaMessage>{m.github_close_and_continue_cli()}</CtaMessage>
  </CtaContentContainer>
</CtaLayoutContainer>
