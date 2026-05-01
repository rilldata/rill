<script lang="ts">
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaNeedHelp from "@rilldata/web-common/components/calls-to-action/CTANeedHelp.svelte";
  import LoadingSpinner from "@rilldata/web-common/components/LoadingSpinner.svelte";

  export let multipleDashboards: boolean = false;
  export let variant: "build" | "publish" = "build";
</script>

<CtaLayoutContainer>
  <CtaContentContainer>
    <LoadingSpinner />
    {#if variant === "publish"}
      <CtaHeader variant="bold">Publishing your project to production...</CtaHeader>
      <p class="text-sm text-fg-secondary -mt-2 max-w-md text-center">
        Setting up your production environment for the first time. This usually
        takes about a minute.
      </p>
    {:else}
      <CtaHeader variant="bold">
        Hang tight! We're building your dashboard{multipleDashboards
          ? "s"
          : ""}...
      </CtaHeader>
    {/if}
    <CtaNeedHelp />
  </CtaContentContainer>
</CtaLayoutContainer>
