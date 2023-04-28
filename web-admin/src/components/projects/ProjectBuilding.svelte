<script lang="ts">
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import CtaContentContainer from "../calls-to-action/CTAContentContainer.svelte";
  import CtaLayoutContainer from "../calls-to-action/CTALayoutContainer.svelte";
  import CtaHeader from "../calls-to-action/CTAHeader.svelte";
  import CtaButton from "../calls-to-action/CTAButton.svelte";
  import CtaMessage from "../calls-to-action/CTAMessage.svelte";

  export let organization: string;
  export let project: string;
</script>

<CtaLayoutContainer>
  <CtaContentContainer maxWidth="400px">
    <div class="h-36">
      <Spinner status={EntityStatus.Running} size="7rem" duration={725} />
    </div>
    <CtaHeader fontSizeClass="text-lg" fontWeightClass="font-semibold"
      >Hang tight! We're building your dashboard...</CtaHeader
    >
    <CtaButton
      variant="primary-outline"
      on:click={() => {
        window.location.href = `/${organization}/${project}`;
      }}
      >View project status
    </CtaButton>
    <CtaMessage fontSizeClass="text-sm">
      Need help? Reach out to us on <a href="http://bit.ly/3jg4IsF">Discord</a>
    </CtaMessage>
  </CtaContentContainer>
</CtaLayoutContainer>
