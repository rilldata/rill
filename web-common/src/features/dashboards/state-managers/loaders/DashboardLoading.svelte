<script lang="ts">
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types.js";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import { LoadingTracker } from "@rilldata/web-common/lib/LoadingTracker";
  import { fade } from "svelte/transition";

  export let isLoading: boolean;
  export let shortLoadDelay: number = 1000;
  export let longLoadDelay: number = 5000;

  const loadingTracker = new LoadingTracker(shortLoadDelay, longLoadDelay);
  $: loadingTracker.updateLoading(isLoading);
  $: ({ loadingForShortTime, loadingForLongTime } = loadingTracker);
</script>

{#if $loadingForShortTime}
  <CtaLayoutContainer>
    <CtaContentContainer>
      <div class="h-36">
        <Spinner status={EntityStatus.Running} size="7rem" duration={725} />
      </div>
      {#if $loadingForLongTime}
        <h1
          class="text-lg font-semibold text-fg-primary text-center"
          transition:fade|local={{ duration: 50 }}
        >
          Loading your dashboard...If this is taking a while, your database may
          be waking up.
        </h1>
      {/if}
    </CtaContentContainer>
  </CtaLayoutContainer>
{/if}
