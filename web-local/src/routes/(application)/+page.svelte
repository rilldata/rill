<script lang="ts">
  import { OnboardingWorkspace } from "@rilldata/web-local/lib/components/workspace";
  // clear any existing active asset
  // TODO: is this still needed?
  // $: dataModelerService.dispatch("setActiveAsset", [
  //   EntityType.Table,
  //   undefined,
  // ]);
</script>

<svelte:head>
  <title>Rill Developer</title>
  <meta content="DESCRIPTION" name="Description" />
</svelte:head>

<OnboardingWorkspace />
