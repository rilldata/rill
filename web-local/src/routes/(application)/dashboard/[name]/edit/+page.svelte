<script lang="ts">
  import { MetricsDefinitionWorkspace } from "@rilldata/web-local/lib/components/workspace";

  export let data;

  $: metricsDefName = data.metricsDefName;
  $: nonStandardError = data.error;
  $: yaml = data.yaml;
</script>

<svelte:head>
  <!-- TODO: add the dashboard name to the title -->
  <title>Rill Developer</title>
</svelte:head>

<MetricsDefinitionWorkspace {metricsDefName} {nonStandardError} {yaml} />
