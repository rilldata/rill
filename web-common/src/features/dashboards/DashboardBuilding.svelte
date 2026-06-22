<script lang="ts">
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaNeedHelp from "@rilldata/web-common/components/calls-to-action/CTANeedHelp.svelte";
  import LoadingSpinner from "@rilldata/web-common/components/LoadingSpinner.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  export let multipleDashboards: boolean = false;
</script>

<CtaLayoutContainer>
  <CtaContentContainer>
    <LoadingSpinner />
    <CtaHeader variant="bold">
      {multipleDashboards
        ? m.dashboards_building_multiple()
        : m.dashboards_building_single()}
    </CtaHeader>
    <CtaNeedHelp />
  </CtaContentContainer>
</CtaLayoutContainer>
