<script lang="ts">
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaNeedHelp from "@rilldata/web-common/components/calls-to-action/CTANeedHelp.svelte";
  import LoadingSpinner from "@rilldata/web-common/components/LoadingSpinner.svelte";

  export let multipleDashboards: boolean = false;
</script>

<CtaLayoutContainer>
  <CtaContentContainer>
    <LoadingSpinner />
    <CtaHeader variant="bold">
      Hang tight! We're building your dashboard{multipleDashboards
        ? "s"
        : ""}...
    </CtaHeader>
    <CtaNeedHelp />
  </CtaContentContainer>
</CtaLayoutContainer>
