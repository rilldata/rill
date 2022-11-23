<script lang="ts">
  import {
    runtimeServiceGetFile,
    useRuntimeServiceGetFile,
  } from "@rilldata/web-common/runtime-client";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { MetricsDefinitionWorkspace } from "@rilldata/web-local/lib/components/workspace";

  export let data;

  $: metricsDefName = data.metricsDefName;
  $: repoId = $runtimeStore.repoId;

  $: dashboardYAML = useRuntimeServiceGetFile(
    repoId,
    `dashboards/${metricsDefName}.yaml`
  );

  $: yaml = $dashboardYAML.data?.blob || "";
  $: nonStandardError = data.error;
</script>

<svelte:head>
  <!-- TODO: add the dashboard name to the title -->
  <title>Rill Developer</title>
</svelte:head>

{#if yaml}
  <MetricsDefinitionWorkspace {metricsDefName} {nonStandardError} {yaml} />
{/if}
