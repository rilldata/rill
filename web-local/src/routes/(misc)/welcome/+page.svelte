<script lang="ts">
  import { goto } from "$app/navigation";
  import ProjectCards from "@rilldata/web-common/features/welcome/ProjectCards.svelte";
  import TitleContent from "@rilldata/web-common/features/welcome/TitleContent.svelte";
  import OnboardingGenerateSampleData from "@rilldata/web-common/features/add-data/OnboardingGenerateSampleData.svelte";
  import ConnectYourDataSmall from "@rilldata/web-common/features/add-data/ConnectYourDataSmall.svelte";

  import { AddDataStep } from "@rilldata/web-common/features/add-data/steps/types.ts";
</script>

<div class="flex size-full overflow-hidden">
  <div class="scroll">
    <div class="wrapper column p-10 2xl:py-16">
      <TitleContent />

      <div class="flex flex-row gap-x-12">
        <ConnectYourDataSmall
          startConnectorSelection={(name) =>
            void goto("/welcome/add-data", {
              state: {
                step: name
                  ? AddDataStep.CreateConnector
                  : AddDataStep.SelectConnector,
                schema: name,
              },
            })}
          onWelcomeScreen
        />
        <OnboardingGenerateSampleData />
      </div>

      <div class="column">
        <ProjectCards />
      </div>
    </div>
  </div>
</div>

<style lang="postcss">
  .scroll {
    @apply size-full overflow-x-hidden overflow-y-auto;
  }

  .wrapper {
    @apply w-full h-fit min-h-screen bg-no-repeat bg-cover;
    background-image: url("/img/welcome-bg-art.jpg");
  }

  :global(.dark) .wrapper {
    background-image: url("/img/welcome-bg-art-dark.jpg");
  }

  .column {
    @apply flex flex-col items-center gap-y-6;
  }
</style>
