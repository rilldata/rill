<script lang="ts">
  import { goto } from "$app/navigation";
  import ProjectCards from "@rilldata/web-common/features/welcome/ProjectCards.svelte";
  import TitleContent from "@rilldata/web-common/features/welcome/TitleContent.svelte";
  import OnboardingGenerateSampleData from "@rilldata/web-common/features/add-data/OnboardingGenerateSampleData.svelte";
  import ConnectYourDataWidget from "@rilldata/web-common/features/add-data/ConnectYourDataWidget.svelte";
  import { onMount } from "svelte";
  import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics.ts";
  import {
    BehaviourEventAction,
    BehaviourEventMedium,
  } from "@rilldata/web-common/metrics/service/BehaviourEventTypes.ts";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-common/metrics/service/MetricsTypes.ts";

  onMount(() => {
    void behaviourEvent?.fireAddDataStepEvent(
      BehaviourEventAction.WelcomePageViewed,
      BehaviourEventMedium.Card,
      MetricsEventSpace.Workspace,
      MetricsEventScreenName.Splash,
      {},
    );
  });
</script>

<div class="my-auto">
  <TitleContent />

  <div class="flex flex-col py-6 gap-y-[28px]">
    <div class="flex flex-row gap-x-12">
      <ConnectYourDataWidget
        startConnectorSelection={(name) =>
          void goto("/welcome/add-data" + (name ? `?schema=${name}` : ""))}
        onWelcomeScreen
      />
      <OnboardingGenerateSampleData />
    </div>

    <p class="text-base font-normal text-fg-secondary text-center">
      Or jump right into an example project.
    </p>

    <ProjectCards />
  </div>
</div>
