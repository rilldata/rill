<script lang="ts">
  import OnboardingWorkspace from "@rilldata/web-common/features/onboarding/OnboardingWorkspace.svelte";
  import WelcomeScreen from "@rilldata/web-common/features/welcome/WelcomeScreen.svelte";
  import type { LayoutData } from "../$types";
  import GeneratingSampleDataMessage from "@rilldata/web-common/features/sample-data/GeneratingSampleDataMessage.svelte";
  import DeveloperChat from "@rilldata/web-common/features/chat/DeveloperChat.svelte";
  import { generatingSampleData } from "@rilldata/web-common/features/sample-data/generate-sample-data.ts";

  export let data: LayoutData;
</script>

<svelte:head>
  <title>Rill Developer</title>
</svelte:head>

<div class="flex size-full overflow-hidden">
  <div class="flex size-full overflow-hidden">
    {#if data.initialized}
      {#if $generatingSampleData}
        <GeneratingSampleDataMessage />
      {:else}
        <OnboardingWorkspace />
      {/if}
    {:else}
      <WelcomeScreen />
    {/if}
  </div>
  <DeveloperChat />
</div>
