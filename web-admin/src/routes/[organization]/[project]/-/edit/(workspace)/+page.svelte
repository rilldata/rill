<script lang="ts">
  import * as m from "@rilldata/web-common/paraglide/messages.js";
  import { GeneratingMessage } from "@rilldata/web-common/components/generating-message";
  import { generatingSampleData } from "@rilldata/web-common/features/sample-data/generate-sample-data.ts";
  import OnboardingWorkspace from "@rilldata/web-common/features/onboarding/OnboardingWorkspace.svelte";
</script>

<div class="flex size-full overflow-hidden bg-surface-subtle">
  {#if $generatingSampleData}
    <GeneratingMessage title={m.workspace_generating_sample_data()} />
  {:else}
    <OnboardingWorkspace />
  {/if}
</div>
