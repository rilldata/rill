<script lang="ts">
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaNeedHelp from "@rilldata/web-common/components/calls-to-action/CTANeedHelp.svelte";
  import LoadingSpinner from "@rilldata/web-common/components/LoadingSpinner.svelte";
</script>

<CtaLayoutContainer>
  <CtaContentContainer>
    <LoadingSpinner />
    <CtaHeader variant="bold">
      Publishing your project to production...
    </CtaHeader>
    <p class="text-sm text-fg-secondary -mt-2 max-w-md text-center">
      Setting up your production environment for the first time. This usually
      takes about a minute.
    </p>
    <CtaNeedHelp />
  </CtaContentContainer>
</CtaLayoutContainer>
