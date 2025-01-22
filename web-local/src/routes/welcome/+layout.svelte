<script lang="ts">
  import ProgressDots from "@rilldata/web-common/features/welcome/wizard/ProgressDots.svelte";
  import type { LayoutData } from "./$types";

  export let data: LayoutData;

  const {
    onboardingState: { getNumberOfSteps, getCurrentStep },
  } = data;

  const numberOfSteps = getNumberOfSteps();
  const currentStep = getCurrentStep();
</script>

<div class="container">
  <slot />
  <span class="absolute bottom-10">
    <ProgressDots
      numberOfDots={$numberOfSteps}
      activeDotIndex={$currentStep - 1}
    />
  </span>
</div>

<style lang="postcss">
  .container {
    @apply max-w-screen-xl mx-auto h-screen;
    @apply mt-12 px-4 pt-6 pb-4;
    @apply flex flex-col gap-y-4 justify-start items-center text-center;
    @apply relative;
  }
</style>
