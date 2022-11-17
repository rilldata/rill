<script lang="ts">
  import { EntityType } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
  import { dataModelerService } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { OnboardingWorkspace } from "@rilldata/web-local/lib/components/workspace";
  // clear any existing active asset
  $: dataModelerService.dispatch("setActiveAsset", [
    EntityType.Table,
    undefined,
  ]);
</script>

<svelte:head>
  <title>Rill Developer</title>
  <meta name="Description" content="DESCRIPTION" />
</svelte:head>

<OnboardingWorkspace />
