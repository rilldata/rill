<script lang="ts">
  import ProjectCards from "@rilldata/web-common/features/welcome/ProjectCards.svelte";
  import TitleContent from "@rilldata/web-common/features/welcome/TitleContent.svelte";
  import OnboardingGenerateSampleData from "@rilldata/web-common/features/add-data/OnboardingGenerateSampleData.svelte";
  import ConnectYourDataWidget from "@rilldata/web-common/features/add-data/ConnectYourDataWidget.svelte";
</script>

<div class="my-auto">
  <TitleContent />

  <div class="flex flex-col py-6 gap-y-[28px]">
    <div class="flex flex-row gap-x-12">
      <ConnectYourDataWidget onWelcomeScreen />
      <OnboardingGenerateSampleData />
    </div>

    <p class="text-base font-normal text-fg-secondary text-center">
      Or jump right into an example project.
    </p>

    <ProjectCards />
  </div>
</div>
