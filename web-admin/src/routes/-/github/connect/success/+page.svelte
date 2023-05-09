<script>
  import GithubSuccess from "@rilldata/web-common/components/icons/GithubSuccess.svelte";
  import CtaContentContainer from "../../../../../components/calls-to-action/CTAContentContainer.svelte";
  import CtaHeader from "../../../../../components/calls-to-action/CTAHeader.svelte";
  import CtaLayoutContainer from "../../../../../components/calls-to-action/CTALayoutContainer.svelte";
  import CtaMessage from "../../../../../components/calls-to-action/CTAMessage.svelte";
</script>

<svelte:head>
  <title>Github connected successfully</title>
</svelte:head>

<CtaLayoutContainer>
  <CtaContentContainer>
    <GithubSuccess />
    <CtaHeader>Github connected successfully</CtaHeader>
    <CtaMessage>Close this page and continue setup in the Rill CLI.</CtaMessage>
  </CtaContentContainer>
</CtaLayoutContainer>
