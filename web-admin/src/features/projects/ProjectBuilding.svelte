<script lang="ts">
  import { page } from "$app/stores";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaNeedHelp from "@rilldata/web-common/components/calls-to-action/CTANeedHelp.svelte";
  import LoadingSpinner from "@rilldata/web-common/components/LoadingSpinner.svelte";
  import { isEmbedPage } from "@rilldata/web-common/layout/navigation/navigation-utils.ts";

  /** When set, shows a branch-specific loading message instead of the generic one. */
  export let branch: string | undefined = undefined;

  const onEmbedPage = isEmbedPage($page);
</script>

<CtaLayoutContainer>
  <CtaContentContainer>
    <LoadingSpinner />
    <CtaHeader variant="bold">
      {#if branch}
        Starting branch deployment...
      {:else}
        Hang tight! We're deploying your project...
      {/if}
    </CtaHeader>
    {#if !onEmbedPage}
      <CtaNeedHelp />
    {/if}
  </CtaContentContainer>
</CtaLayoutContainer>
