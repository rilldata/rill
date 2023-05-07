<script lang="ts">
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import CtaButton from "../../../../web-common/src/components/calls-to-action/CTAButton.svelte";
  import CtaContentContainer from "../../../../web-common/src/components/calls-to-action/CTAContentContainer.svelte";
  import CtaHeader from "../../../../web-common/src/components/calls-to-action/CTAHeader.svelte";
  import CtaLayoutContainer from "../../../../web-common/src/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaNeedHelp from "../../../../web-common/src/components/calls-to-action/CTANeedHelp.svelte";

  export let organization: string;
  export let project: string;
</script>

<CtaLayoutContainer>
  <CtaContentContainer>
    <div class="h-36">
      <Spinner status={EntityStatus.Running} size="7rem" duration={725} />
    </div>
    <CtaHeader variant="bold"
      >Hang tight! We're building your dashboard...</CtaHeader
    >
    <CtaButton
      variant="primary-outline"
      on:click={() => {
        window.location.href = `/${organization}/${project}`;
      }}
      >View project status
    </CtaButton>
    <CtaNeedHelp />
  </CtaContentContainer>
</CtaLayoutContainer>
