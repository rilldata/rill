<script lang="ts">
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types.js";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import { fade } from "svelte/transition";

  export let loadingForShortTime: boolean;
  export let loadingForLongTime: boolean;
</script>

{#if loadingForShortTime}
  <CtaLayoutContainer>
    <CtaContentContainer>
      <div class="h-36">
        <Spinner status={EntityStatus.Running} size="7rem" duration={725} />
      </div>
      {#if loadingForLongTime}
        <h1
          class="text-lg font-semibold text-gray-800 text-center"
          transition:fade|local={{ duration: 50 }}
        >
          Loading your dashboard...If this is taking a while, your database may
          be waking up.
        </h1>
      {/if}
    </CtaContentContainer>
  </CtaLayoutContainer>
{/if}
