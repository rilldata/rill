<!-- This page is for cases when user authorised the github app on another github account which doesn't have access to the repo  -->
<script lang="ts">
  import { redirectToGithubLogin } from "@rilldata/web-admin/client/redirect-utils";
  import GithubRepoInline from "@rilldata/web-admin/features/projects/github/GithubRepoInline.svelte";
  import GithubUserInline from "@rilldata/web-admin/features/projects/github/GithubUserInline.svelte";
  import CtaButton from "@rilldata/web-common/components/calls-to-action/CTAButton.svelte";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaMessage from "@rilldata/web-common/components/calls-to-action/CTAMessage.svelte";
  import GithubFail from "@rilldata/web-common/components/icons/GithubFail.svelte";

  const urlParams = new URLSearchParams(window.location.search);
  const remote = urlParams.get("remote");
  const redirect = urlParams.get("redirect");
  const githubUsername = urlParams.get("githubUsername");
</script>

<svelte:head>
  <title>Could not connect to GitHub</title>
</svelte:head>

<CtaLayoutContainer>
  <CtaContentContainer>
    <GithubFail />
    <CtaHeader>Could not connect to GitHub</CtaHeader>
    <CtaMessage>
      Your authorized GitHub account <GithubUserInline {githubUsername} />
      does not have access to <GithubRepoInline gitRemote={remote} />.
    </CtaMessage>
    <CtaMessage>
      Click the button below to re-authorize/authorize another account.
    </CtaMessage>
    <CtaButton
      variant="primary"
      onClick={() => redirectToGithubLogin(remote, redirect, "auth")}
    >
      Connect to GitHub
    </CtaButton>
  </CtaContentContainer>
</CtaLayoutContainer>
