<script lang="ts">
  import CustomChartRenderer from "@rilldata/web-common/components/vega/CustomChartRenderer.svelte";
  import type { CustomChartComponent } from "./index";

  export let component: CustomChartComponent;

  $: ({ specStore } = component);

  $: ({ metrics_sql, vega_spec } = $specStore);
</script>

<CustomChartRenderer
  name={component.id}
  spec={vega_spec}
  metricsSQL={metrics_sql}
  showDataTable
/>
