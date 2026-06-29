<script>
  import RillLogoSquareNegative from "@rilldata/web-common/components/icons/RillLogoSquareNegative.svelte";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaMessage from "@rilldata/web-common/components/calls-to-action/CTAMessage.svelte";
  import * as m from "@rilldata/web-common/paraglide/messages.js";
</script>

<svelte:head>
  <title>{m.auth_cli_success_title()}</title>
</svelte:head>

<CtaLayoutContainer>
  <CtaContentContainer>
    <RillLogoSquareNegative size="84px" />
    <CtaHeader>{m.auth_cli_success_title()}</CtaHeader>
    <CtaMessage>{m.auth_cli_success_message()}</CtaMessage>
  </CtaContentContainer>
</CtaLayoutContainer>
