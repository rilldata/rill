<script lang="ts">
  import AddDataManager from "@rilldata/web-common/features/add-data/manager/AddDataManager.svelte";
  import { AddDataStep } from "@rilldata/web-common/features/add-data/manager/steps/types.ts";
  import { WelcomeStatus } from "@rilldata/web-common/features/welcome/status.ts";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-common/metrics/service/MetricsTypes.ts";
  import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes.ts";

  export let data;

  let addDataStep: AddDataStep = AddDataStep.SelectConnector;

  $: isImportStep = addDataStep === AddDataStep.Import;
</script>

<div class="my-auto">
  {#if !isImportStep}
    <div class="text-base font-semibold text-fg-secondary">Getting started</div>
    <div class="text-3xl font-bold text-fg-accent">Connect your data</div>
  {/if}
  <div class="w-fit h-fit mt-4">
    <AddDataManager
      config={{
        welcomeScreen: true,
        medium: BehaviourEventMedium.Card,
        space: MetricsEventSpace.Workspace,
        screen: MetricsEventScreenName.Splash,
      }}
      initSchema={data.schema}
      onStepChange={(step) => (addDataStep = step)}
      onClose={() => window.history.back()}
      onDone={() => WelcomeStatus.set(false)}
    />
  </div>
</div>
