<script lang="ts">
  import { useQuery } from "@sveltestack/svelte-query";
  import { fetchWrapper } from "$lib/util/fetchWrapper";

  const metricsDefinitions = useQuery("metrics", () =>
    fetchWrapper("metrics", "GET")
  );
  $: modifiedDefinitions =
    $metricsDefinitions?.data?.filter((metric) =>
      metric.metricDefLabel.endsWith("-")
    ) ?? [];
</script>

<div>
  <h3>Modified Metrics Definitions</h3>
  {#each modifiedDefinitions as metric (metric.id)}
    <div>{metric.metricDefLabel}</div>
  {/each}
</div>
