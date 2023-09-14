<script lang="ts">
  import { goto } from "$app/navigation";
  import CtaButton from "@rilldata/web-common/components/calls-to-action/CTAButton.svelte";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaMessage from "@rilldata/web-common/components/calls-to-action/CTAMessage.svelte";
  import CtaNeedHelp from "@rilldata/web-common/components/calls-to-action/CTANeedHelp.svelte";

  export let statusCode: number;
  export let header: string;
  export let body: string;
</script>

<CtaLayoutContainer>
  <CtaContentContainer>
    <h1
      class="text-8xl font-extrabold bg-gradient-to-b from-[#CBD5E1] to-[#E2E8F0] text-transparent bg-clip-text"
    >
      {statusCode}
    </h1>
    <h2 class="text-lg font-semibold">{header}</h2>
    <CtaMessage>
      {body}
    </CtaMessage>
    <CtaButton variant="primary-outline" on:click={() => goto("/")}>
      Back to home
    </CtaButton>
    <CtaNeedHelp />
  </CtaContentContainer>
</CtaLayoutContainer>
