<script lang="ts">
  import { GeneratingMessage } from "@rilldata/web-common/components/generating-message";
  import { generatingSampleData } from "@rilldata/web-common/features/sample-data/generate-sample-data.ts";
  import { generatingCanvas } from "@rilldata/web-common/features/canvas/ai-generation/generateCanvas";
  import DeveloperChat from "@rilldata/web-common/features/chat/DeveloperChat.svelte";
  import OnboardingWorkspace from "@rilldata/web-common/features/onboarding/OnboardingWorkspace.svelte";
  import type { LayoutData } from "../$types";

  export let data: LayoutData;
</script>

<svelte:head>
  <title>Rill Developer</title>
</svelte:head>

<div class="flex size-full overflow-hidden">
  <div class="flex size-full overflow-hidden">
    {#if data.initialized}
      {#if $generatingSampleData}
        <GeneratingMessage title="Generating your sample data..." />
      {:else if $generatingCanvas}
        <GeneratingMessage title="Generating your Canvas dashboard..." />
      {:else}
        <OnboardingWorkspace />
      {/if}
    {/if}
  </div>
  <DeveloperChat />
</div>
