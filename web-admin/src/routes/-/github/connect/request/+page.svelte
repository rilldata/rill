<script lang="ts">
  import GithubRepoInline from "@rilldata/web-admin/features/projects/github/GithubRepoInline.svelte";
  import CodeBlockInline from "@rilldata/web-common/components/calls-to-action/CodeBlockInline.svelte";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaMessage from "@rilldata/web-common/components/calls-to-action/CTAMessage.svelte";
  import KeyboardKey from "@rilldata/web-common/components/calls-to-action/KeyboardKey.svelte";
  import Github from "@rilldata/web-common/components/icons/Github.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  const remote = new URLSearchParams(window.location.search).get("remote");
</script>

<svelte:head>
  <title>{m.github_access_requested_title()}</title>
</svelte:head>

<CtaLayoutContainer>
  <CtaContentContainer>
    <Github className="w-10 h-10 text-fg-primary" />
    <CtaHeader>{m.github_connect_to_github()}</CtaHeader>
    <CtaMessage>
      {m.github_requested_access()} <GithubRepoInline gitRemote={remote} />{m.github_close_page_now()}
    </CtaMessage>
    <CtaMessage>
      {m.github_cli_polling()}
      <KeyboardKey label="Control" /> +
      <KeyboardKey label="C" /> {m.github_run_deploy_again()}
    </CtaMessage>
  </CtaContentContainer>
</CtaLayoutContainer>
