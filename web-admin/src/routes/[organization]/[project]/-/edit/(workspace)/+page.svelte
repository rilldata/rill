<script lang="ts">
  import { GeneratingMessage } from "@rilldata/web-common/components/generating-message";
  import { generatingSampleData } from "@rilldata/web-common/features/sample-data/generate-sample-data.ts";
  import OnboardingWorkspace from "@rilldata/web-common/features/onboarding/OnboardingWorkspace.svelte";
</script>

<div class="flex size-full overflow-hidden bg-surface-subtle">
  {#if $generatingSampleData}
    <GeneratingMessage title="Generating your sample data..." />
  {:else}
    <OnboardingWorkspace />
  {/if}
</div>
