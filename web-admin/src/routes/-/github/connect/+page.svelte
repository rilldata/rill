<script lang="ts">
  import GithubRepoInline from "@rilldata/web-admin/features/projects/github/GithubRepoInline.svelte";
  import CtaButton from "@rilldata/web-common/components/calls-to-action/CTAButton.svelte";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaMessage from "@rilldata/web-common/components/calls-to-action/CTAMessage.svelte";
  import Github from "@rilldata/web-common/components/icons/Github.svelte";

  const urlParams = new URLSearchParams(window.location.search);
  const redirectURL = urlParams.get("redirect");
  const remote = new URL(decodeURIComponent(redirectURL)).searchParams.get(
    "remote",
  );
</script>

<svelte:head>
  <title>Connect to GitHub</title>
</svelte:head>

<CtaLayoutContainer>
  <CtaContentContainer>
    <Github className="w-10 h-10 text-gray-900" />
    <CtaHeader>Connect to GitHub</CtaHeader>
    <CtaMessage>
      Rill projects deploy continuously when you push changes to GitHub.
    </CtaMessage>
    {#if remote}
      <CtaMessage>
        Please grant access to your repository<br /><GithubRepoInline
          githubUrl={remote}
        />
      </CtaMessage>
    {/if}
    <div class="mt-4 w-full flex justify-center">
      <CtaButton variant="primary" href={redirectURL} rel="external">
        Connect to GitHub
      </CtaButton>
    </div>
  </CtaContentContainer>
</CtaLayoutContainer>
