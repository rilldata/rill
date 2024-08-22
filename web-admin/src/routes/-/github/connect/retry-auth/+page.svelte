<!-- This page is for cases when user authorised the github app on another github account which doesn't have access to the repo  -->
<script lang="ts">
  import { goto } from "$app/navigation";
  import { createAdminServiceGetCurrentUser } from "@rilldata/web-admin/client";
  import { CANONICAL_ADMIN_URL } from "@rilldata/web-admin/client/http-client";
  import CtaButton from "@rilldata/web-common/components/calls-to-action/CTAButton.svelte";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaMessage from "@rilldata/web-common/components/calls-to-action/CTAMessage.svelte";
  import GithubFail from "@rilldata/web-common/components/icons/GithubFail.svelte";
  import GithubRepoInline from "@rilldata/web-admin/features/projects/github/GithubRepoInline.svelte";
  import GithubUserInline from "@rilldata/web-admin/features/projects/github/GithubUserInline.svelte";

  const urlParams = new URLSearchParams(window.location.search);
  const remote = urlParams.get("remote");
  const githubUsername = urlParams.get("githubUsername");

  const user = createAdminServiceGetCurrentUser({
    query: {
      onSuccess: (data) => {
        if (!data.user) {
          goto(
            `${CANONICAL_ADMIN_URL}/auth/login?redirect=${window.location.href}`,
          );
        }
      },
    },
  });
</script>

<svelte:head>
  <title>Could not connect to GitHub</title>
</svelte:head>

{#if $user.data && $user.data.user}
  <CtaLayoutContainer>
    <CtaContentContainer>
      <GithubFail />
      <CtaHeader>Could not connect to GitHub</CtaHeader>
      <CtaMessage>
        Your authorized GitHub account <GithubUserInline {githubUsername} />
        does not have access to <GithubRepoInline githubUrl={remote} />.
      </CtaMessage>
      <CtaMessage>
        Click the button below to re-authorize/authorize another account.
      </CtaMessage>
      <CtaButton
        variant="primary"
        href={encodeURI(
          CANONICAL_ADMIN_URL + "/github/auth/login?remote=" + remote,
        )}
      >
        Connect to GitHub
      </CtaButton>
    </CtaContentContainer>
  </CtaLayoutContainer>
{/if}
