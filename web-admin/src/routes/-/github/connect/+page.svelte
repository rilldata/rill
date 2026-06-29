<script lang="ts">
  import GithubRepoInline from "@rilldata/web-admin/features/projects/github/GithubRepoInline.svelte";
  import CtaButton from "@rilldata/web-common/components/calls-to-action/CTAButton.svelte";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaMessage from "@rilldata/web-common/components/calls-to-action/CTAMessage.svelte";
  import Github from "@rilldata/web-common/components/icons/Github.svelte";
  import * as m from "@rilldata/web-common/paraglide/messages.js";

  const urlParams = new URLSearchParams(window.location.search);
  const redirectURL = urlParams.get("redirect");
  const remote = new URL(decodeURIComponent(redirectURL)).searchParams.get(
    "remote",
  );
</script>

<svelte:head>
  <title>{m.github_connect_to_github()}</title>
</svelte:head>

<CtaLayoutContainer>
  <CtaContentContainer>
    <Github className="w-10 h-10 text-fg-primary" />
    <CtaHeader>{m.github_connect_to_github()}</CtaHeader>
    <CtaMessage>
      {m.github_deploy_continuously()}
    </CtaMessage>
    {#if remote}
      <CtaMessage>
        {m.github_grant_access_to_repo()}<br /><GithubRepoInline
          gitRemote={remote}
        />
      </CtaMessage>
    {/if}
    <div class="mt-4 w-full flex justify-center">
      <CtaButton variant="primary" href={redirectURL} rel="external">
        {m.github_connect_to_github()}
      </CtaButton>
    </div>
  </CtaContentContainer>
</CtaLayoutContainer>
