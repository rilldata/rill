<script lang="ts">
  import {
    useQuery,
    useMutation,
    useQueryClient,
  } from "@sveltestack/svelte-query";
  import type { MetricsDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
  import { fetchWrapper } from "$lib/util/fetchWrapper";
  import UpdatedMetrics from "./UpdatedMetrics.svelte";
  import QueriedExplore from "./QueriedExplore.svelte";

  const queryClient = useQueryClient();

  const metricsDefinitions = useQuery<Array<MetricsDefinitionEntity>>(
    "metrics",
    () => fetchWrapper("metrics", "GET")
  );

  const updateMetricsDefinition = useMutation(
    (metrics: MetricsDefinitionEntity) =>
      fetchWrapper(`metrics/${metrics.id}`, "POST", metrics),
    {
      onSuccess: () => {
        queryClient.invalidateQueries("metrics");
      },
    }
  );

  // Simple update call to demonstrate caching
  const updateMetricsDef = (metrics: MetricsDefinitionEntity) => {
    $updateMetricsDefinition.mutate({
      ...metrics,
      metricDefLabel: metrics.metricDefLabel.endsWith("-")
        ? metrics.metricDefLabel.replace(/-$/, "")
        : metrics.metricDefLabel + "-",
    });
  };
</script>

<div class="fixed" style="width: 30%;">
  {#if $metricsDefinitions.isLoading}
    <span>Loading...</span>
  {:else if $metricsDefinitions.error}
    <span>An error has occurred: {$metricsDefinitions.error.message}</span>
  {:else}
    {#each $metricsDefinitions.data as metrics (metrics.id)}
      <div>
        <h1>{metrics.metricDefLabel}</h1>
        <button on:click={() => updateMetricsDef(metrics)}>Update</button>
      </div>
    {/each}
  {/if}
  <UpdatedMetrics />
</div>
<div class="fixed" style="left: 30%;"><QueriedExplore /></div>
