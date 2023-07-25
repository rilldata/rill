<script>
  import RillLogoSquareNegative from "@rilldata/web-common/components/icons/RillLogoSquareNegative.svelte";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaMessage from "@rilldata/web-common/components/calls-to-action/CTAMessage.svelte";
</script>

<svelte:head>
  <title>Successfully authorized Rill CLI</title>
</svelte:head>

<CtaLayoutContainer>
  <CtaContentContainer>
    <RillLogoSquareNegative size="84px" />
    <CtaHeader>Successfully authorized Rill CLI</CtaHeader>
    <CtaMessage>Close this page and continue in the Rill CLI.</CtaMessage>
  </CtaContentContainer>
</CtaLayoutContainer>
