<!-- When we navigate users to install page. 
  We can't control the repo users install the github app on and they can end up installing the app on another repo.
  This page is for showing them the message that github app is installed on another repo than they need to reinstall app on right repo.  -->
<script lang="ts">
  import { redirectToGithubLogin } from "@rilldata/web-admin/client/redirect-utils";
  import GithubRepoInline from "@rilldata/web-admin/features/projects/github/GithubRepoInline.svelte";
  import CtaButton from "@rilldata/web-common/components/calls-to-action/CTAButton.svelte";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaMessage from "@rilldata/web-common/components/calls-to-action/CTAMessage.svelte";
  import KeyboardKey from "@rilldata/web-common/components/calls-to-action/KeyboardKey.svelte";
  import GithubFail from "@rilldata/web-common/components/icons/GithubFail.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  const urlParams = new URLSearchParams(window.location.search);
  const remote = urlParams.get("remote");
  const redirect = urlParams.get("redirect");
</script>

<svelte:head>
  <title>{m.github_could_not_connect()}</title>
</svelte:head>

<CtaLayoutContainer>
  <CtaContentContainer>
    <GithubFail />
    <CtaHeader>{m.github_could_not_connect()}</CtaHeader>
    <CtaMessage>
      {m.github_did_not_grant_access()} <GithubRepoInline
        gitRemote={remote}
      />
    </CtaMessage>
    <CtaMessage>
      {m.github_click_to_retry()}
      <KeyboardKey label="Control" /> + <KeyboardKey label="C" /> {m.github_cancel_in_cli()}
    </CtaMessage>
    <CtaButton
      variant="primary"
      onClick={() => redirectToGithubLogin(remote, redirect, "connect")}
    >
      {m.github_connect_to_github()}
    </CtaButton>
  </CtaContentContainer>
</CtaLayoutContainer>
