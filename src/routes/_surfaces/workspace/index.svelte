<script lang="ts">
  import { getContext } from "svelte";
  import ModelView from "./Model.svelte";
  import ModelWorkspaceHeader from "./ModelWorkspaceHeader.svelte";
  // import MetricsDefWorkspaceHeader from "./metrics-def/MetricsDefWorkspaceHeader.svelte";
  // import MetricsDefWorkspace from "./metrics-def/MetricsDefWorkspace.svelte";

  import type { ApplicationStore } from "$lib/application-state-stores/application-store";

  import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";

  let MetricsDefWorkspaceHeader = null;
  let MetricsDefWorkspace = null;

  if (import.meta.env.VITE_USE_METRICS_DEF === `use`) {
    (async () => {
      MetricsDefWorkspaceHeader = (
        await import("./metrics-def/MetricsDefWorkspaceHeader.svelte")
      ).default;
      MetricsDefWorkspace = (
        await import("./metrics-def/MetricsDefWorkspace.svelte")
      ).default;
    })();
  }

  const rillAppStore = getContext("rill:app:store") as ApplicationStore;

  $: useModelWorkspace = $rillAppStore?.activeEntity?.type === EntityType.Model;
  $: useMetricsDefWorkspace =
    $rillAppStore?.activeEntity?.type === EntityType.MetricsDef;
  $: activeEntityID = $rillAppStore?.activeEntity?.id;
</script>

{#if useModelWorkspace}
  <ModelWorkspaceHeader />
  <ModelView />
{:else if useMetricsDefWorkspace && MetricsDefWorkspaceHeader !== null && MetricsDefWorkspace !== null}
  <MetricsDefWorkspaceHeader metricsDefId={activeEntityID} />
  <MetricsDefWorkspace />
{/if}
