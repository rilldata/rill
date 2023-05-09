<!-- When we navigate users to install page. 
  We can't control the repo users install the github app on and they can end up installing the app on another repo.
  This page is for showing them the message that github app is installed on another repo than they need to reinstall app on right repo.  -->
<script lang="ts">
  import { goto } from "$app/navigation";
  import { createAdminServiceGetCurrentUser } from "@rilldata/web-admin/client";
  import { ADMIN_URL } from "@rilldata/web-admin/client/http-client";
  import GithubFail from "@rilldata/web-common/components/icons/GithubFail.svelte";
  import CtaButton from "@rilldata/web-common/components/calls-to-action/CTAButton.svelte";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaMessage from "@rilldata/web-common/components/calls-to-action/CTAMessage.svelte";
  import KeyboardKey from "@rilldata/web-common/components/calls-to-action/KeyboardKey.svelte";
  import GithubRepoInline from "../../../../../components/projects/GithubRepoInline.svelte";

  const remote = new URLSearchParams(window.location.search).get("remote");
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
      ADMIN_URL + "/github/connect?remote=" + remote
    );
  }
</script>

<svelte:head>
  <title>Connect to Github</title>
</svelte:head>

{#if $user.data && $user.data.user}
  <CtaLayoutContainer>
    <CtaContentContainer>
      <GithubFail />
      <CtaHeader>Connect to Github</CtaHeader>
      <CtaMessage>
        It looks like you did not grant access to the desired repository at <GithubRepoInline
          githubUrl={remote}
        />.
      </CtaMessage>
      <CtaMessage>
        Click the button below to retry. (Or if this was intentional, press
        <KeyboardKey label="Control" /> + <KeyboardKey label="C" /> in the CLI to
        cancel the connect request.)
      </CtaMessage>
      <CtaButton variant="primary" on:click={handleGoToGithub}>
        Connect to Github
      </CtaButton>
    </CtaContentContainer>
  </CtaLayoutContainer>
{/if}
