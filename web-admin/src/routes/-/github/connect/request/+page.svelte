<script lang="ts">
  import GithubRepoInline from "@rilldata/web-admin/features/projects/github/GithubRepoInline.svelte";
  import CodeBlockInline from "@rilldata/web-common/components/calls-to-action/CodeBlockInline.svelte";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaMessage from "@rilldata/web-common/components/calls-to-action/CTAMessage.svelte";
  import KeyboardKey from "@rilldata/web-common/components/calls-to-action/KeyboardKey.svelte";
  import Github from "@rilldata/web-common/components/icons/Github.svelte";

  const remote = new URLSearchParams(window.location.search).get("remote");
</script>

<svelte:head>
  <title>GitHub access requested</title>
</svelte:head>

<CtaLayoutContainer>
  <CtaContentContainer>
    <Github className="w-10 h-10 text-gray-900" />
    <CtaHeader>Connect to GitHub</CtaHeader>
    <CtaMessage>
      You requested access to <GithubRepoInline githubUrl={remote} />. You can
      close this page now.
    </CtaMessage>
    <CtaMessage>
      The CLI will keep polling until GitHub access has been granted by an
      admin. You can stop polling by pressing <KeyboardKey label="Control" /> +
      <KeyboardKey label="C" /> and run <CodeBlockInline code="rill deploy" />
      again once access has been granted.
    </CtaMessage>
  </CtaContentContainer>
</CtaLayoutContainer>
