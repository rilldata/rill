<script lang="ts">
  import { DashboardWorkspace } from "@rilldata/web-common/features/dashboards";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/dashboard-stores";
  import { base64ToProto } from "@rilldata/web-common/features/dashboards/proto-state/fromProto";
  import { fromProto } from "@rilldata/web-common/features/dashboards/proto-state/fromProto.js";
  import { onMount, tick } from "svelte";

  export let data;

  $: metricViewName = data.metricViewName;

  onMount(async () => {
    await tick();
    const state = new URL(location.href).searchParams.get("state");
    const [filters, selectedTimeRage] = fromProto(base64ToProto(state));
    metricsExplorerStore.create(metricViewName, filters, selectedTimeRage);
  });
</script>

<svelte:head>
  <title>Rill Developer | {metricViewName}</title>
</svelte:head>

<DashboardWorkspace {metricViewName} />
