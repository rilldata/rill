<script lang="ts">
  import { goto } from "$app/navigation";
  import { createAdminServiceGetCurrentUser } from "@rilldata/web-admin/client";
  import { ADMIN_URL } from "@rilldata/web-admin/client/http-client";
  import Github from "@rilldata/web-common/components/icons/Github.svelte";
  import CtaButton from "../../../../components/calls-to-action/CTAButton.svelte";
  import CtaContentContainer from "../../../../components/calls-to-action/CTAContentContainer.svelte";
  import CtaHeader from "../../../../components/calls-to-action/CTAHeader.svelte";
  import CtaLayoutContainer from "../../../../components/calls-to-action/CTALayoutContainer.svelte";
  import CtaMessage from "../../../../components/calls-to-action/CTAMessage.svelte";
  import GithubRepoInline from "../../../../components/projects/GithubRepoInline.svelte";

  const urlParams = new URLSearchParams(window.location.search);
  const redirectURL = urlParams.get("redirect");
  const remote = new URL(decodeURIComponent(redirectURL)).searchParams.get(
    "remote"
  );
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
    window.location.href = redirectURL;
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
        Please grant read-only access to your repository <GithubRepoInline
          githubUrl={remote}
        />
      </CtaMessage>
      <div class="mt-4 w-full">
        <CtaButton variant="primary" on:click={handleGoToGithub}
          >Connect to Github</CtaButton
        >
      </div>
    </CtaContentContainer>
  </CtaLayoutContainer>
{/if}
