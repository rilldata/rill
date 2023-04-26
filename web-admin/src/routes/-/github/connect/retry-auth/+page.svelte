<!-- This page is for cases when user authorised the github app on another github account which doesn't have access to the repo  -->
<script lang="ts">
  import { goto } from "$app/navigation";
  import { createAdminServiceGetCurrentUser } from "@rilldata/web-admin/client";
  import { ADMIN_URL } from "@rilldata/web-admin/client/http-client";
  import Github from "@rilldata/web-common/components/icons/Github.svelte";
  import CtaButton from "../../../../../components/calls-to-action/CTAButton.svelte";
  import CtaContentContainer from "../../../../../components/calls-to-action/CTAContentContainer.svelte";
  import CtaHeader from "../../../../../components/calls-to-action/CTAHeader.svelte";
  import CtaLayoutContainer from "../../../../../components/calls-to-action/CTALayoutContainer.svelte";
  import CtaMessage from "../../../../../components/calls-to-action/CTAMessage.svelte";
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
  <title>Connect to Github</title>
</svelte:head>

{#if $user.data && $user.data.user}
  <CtaLayoutContainer>
    <CtaContentContainer>
      <Github className="w-10 h-10 text-gray-900" />
      <CtaHeader>Connect to Github</CtaHeader>
      <CtaMessage>
        Your authorised github user <GithubUserInline {githubUsername} />
        is not a collaborator to repo <GithubRepoInline githubUrl={remote} />.
      </CtaMessage>
      <CtaMessage>
        Click the button below to re-authorise/authorise another account.
      </CtaMessage>
      <CtaButton variant="primary" on:click={handleGoToGithub}>
        Connect to Github
      </CtaButton>
    </CtaContentContainer>
  </CtaLayoutContainer>
{/if}
