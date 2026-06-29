<script lang="ts">
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaNeedHelp from "@rilldata/web-common/components/calls-to-action/CTANeedHelp.svelte";
  import RadixLarge from "@rilldata/web-common/components/typography/RadixLarge.svelte";
  import Subheading from "@rilldata/web-common/components/typography/Subheading.svelte";
  import * as m from "@rilldata/web-common/paraglide/messages.js";
</script>

<CtaLayoutContainer>
  <CtaContentContainer>
    <div class="flex flex-col gap-y-2">
      <RadixLarge>{m.auth_thanks_for_signing_up()}</RadixLarge>
      <Subheading>{m.auth_check_email_to_verify()}</Subheading>
    </div>
    <CtaNeedHelp />
  </CtaContentContainer>
</CtaLayoutContainer>
