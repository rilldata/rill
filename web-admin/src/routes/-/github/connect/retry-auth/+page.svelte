<!-- This page is for cases when user authorised the github app on another github account which doesn't have access to the repo  -->
<script lang="ts">
  import { goto } from "$app/navigation";
  import { createAdminServiceGetCurrentUser } from "@rilldata/web-admin/client";
  import { ADMIN_URL } from "@rilldata/web-admin/client/http-client";
  import CtaButton from "@rilldata/web-common/components/calls-to-action/CTAButton.svelte";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaMessage from "@rilldata/web-common/components/calls-to-action/CTAMessage.svelte";
  import GithubFail from "@rilldata/web-common/components/icons/GithubFail.svelte";
  import GithubRepoInline from "../../../../../components/projects/GithubRepoInline.svelte";
  import GithubUserInline from "../../../../../components/projects/GithubUserInline.svelte";

  const urlParams = new URLSearchParams(window.location.search);
  const remote = urlParams.get("remote");
  const githubUsername = urlParams.get("githubUsername");
  const user = createAdminServiceGetCurrentUser({
    query: {
      onSuccess: (data) => {
        if (!data.user) {
          goto(`${ADMIN_URL}/auth/login?redirect=${window.location.href}`);
        }
      },
    },
  });

  function handleGoToGithub() {
    window.location.href = encodeURI(
      ADMIN_URL + "/github/auth/login?remote=" + remote
    );
  }
</script>

<svelte:head>
  <title>Could not connect to Github</title>
</svelte:head>

{#if $user.data && $user.data.user}
  <CtaLayoutContainer>
    <CtaContentContainer>
      <GithubFail />
      <CtaHeader>Could not connect to Github</CtaHeader>
      <CtaMessage>
        Your authorized Github account <GithubUserInline {githubUsername} />
        does not have access to <GithubRepoInline githubUrl={remote} />.
      </CtaMessage>
      <CtaMessage>
        Click the button below to re-authorize/authorize another account.
      </CtaMessage>
      <CtaButton variant="primary" on:click={handleGoToGithub}>
        Connect to Github
      </CtaButton>
    </CtaContentContainer>
  </CtaLayoutContainer>
{/if}
