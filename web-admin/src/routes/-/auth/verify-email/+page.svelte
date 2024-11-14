<script lang="ts">
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaNeedHelp from "@rilldata/web-common/components/calls-to-action/CTANeedHelp.svelte";
  import RadixLarge from "@rilldata/web-common/components/typography/RadixLarge.svelte";
  import Subheading from "@rilldata/web-common/components/typography/Subheading.svelte";
</script>

<CtaLayoutContainer>
  <CtaContentContainer>
    <div class="flex flex-col gap-y-2">
      <RadixLarge>Thanks for signing up!</RadixLarge>
      <Subheading>Please check your email to verify your account.</Subheading>
    </div>
    <CtaNeedHelp />
  </CtaContentContainer>
</CtaLayoutContainer>
