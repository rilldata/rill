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

  const urlParams = new URLSearchParams(window.location.search);
  const remote = urlParams.get("remote");
  const redirect = urlParams.get("redirect");
</script>

<svelte:head>
  <title>Could not connect to GitHub</title>
</svelte:head>

<CtaLayoutContainer>
  <CtaContentContainer>
    <GithubFail />
    <CtaHeader>Could not connect to GitHub</CtaHeader>
    <CtaMessage>
      It looks like you did not grant access to the desired repository at <GithubRepoInline
        gitRemote={remote}
      />.
    </CtaMessage>
    <CtaMessage>
      Click the button below to retry. (Or if this was intentional, press
      <KeyboardKey label="Control" /> + <KeyboardKey label="C" /> in the CLI to cancel
      the connect request.)
    </CtaMessage>
    <CtaButton
      variant="primary"
      onClick={() => redirectToGithubLogin(remote, redirect, "connect")}
    >
      Connect to GitHub
    </CtaButton>
  </CtaContentContainer>
</CtaLayoutContainer>
