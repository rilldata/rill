<script lang="ts">
  import { goto } from "$app/navigation";
  import ProjectCards from "@rilldata/web-common/features/welcome/ProjectCards.svelte";
  import TitleContent from "@rilldata/web-common/features/welcome/TitleContent.svelte";
  import OnboardingGenerateSampleData from "@rilldata/web-common/features/add-data/OnboardingGenerateSampleData.svelte";
  import ConnectYourDataWidget from "@rilldata/web-common/features/add-data/ConnectYourDataWidget.svelte";

  import { AddDataStep } from "@rilldata/web-common/features/add-data/steps/types.ts";
</script>

<TitleContent />

<div class="flex flex-row gap-x-12">
  <ConnectYourDataWidget
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

<style lang="postcss">
  .column {
    @apply flex flex-col items-center gap-y-6;
  }
</style>
