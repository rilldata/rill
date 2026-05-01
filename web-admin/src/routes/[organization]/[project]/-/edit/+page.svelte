<script lang="ts">
  import { GeneratingMessage } from "@rilldata/web-common/components/generating-message";
  import { generatingSampleData } from "@rilldata/web-common/features/sample-data/generate-sample-data.ts";
  import { generatingCanvas } from "@rilldata/web-common/features/canvas/ai-generation/generateCanvas";
  import OnboardingWorkspace from "@rilldata/web-common/features/onboarding/OnboardingWorkspace.svelte";
</script>

<div class="flex size-full overflow-hidden bg-surface-subtle">
  {#if $generatingSampleData}
    <GeneratingMessage title="Generating your sample data..." />
  {:else if $generatingCanvas}
    <GeneratingMessage title="Generating your Canvas dashboard..." />
  {:else}
    <OnboardingWorkspace />
  {/if}
</div>
