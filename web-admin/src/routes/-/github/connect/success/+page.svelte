<script>
  import Github from "@rilldata/web-common/components/icons/Github.svelte";
  import CtaContentContainer from "../../../../../../../web-common/src/components/calls-to-action/CTAContentContainer.svelte";
  import CtaHeader from "../../../../../../../web-common/src/components/calls-to-action/CTAHeader.svelte";
  import CtaLayoutContainer from "../../../../../../../web-common/src/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaMessage from "../../../../../../../web-common/src/components/calls-to-action/CTAMessage.svelte";
</script>

<svelte:head>
  <title>Github connected successfully</title>
</svelte:head>

<CtaLayoutContainer>
  <CtaContentContainer>
    <Github className="w-10 h-10 text-gray-900" />
    <CtaHeader>Github connected successfully</CtaHeader>
    <CtaMessage>Close this page and continue setup in the Rill CLI.</CtaMessage>
  </CtaContentContainer>
</CtaLayoutContainer>
