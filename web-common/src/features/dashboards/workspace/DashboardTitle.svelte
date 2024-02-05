<script lang="ts">
  import { runtime } from "../../../runtime-client/runtime-store";
  import { useMetricsView } from "../selectors";

  export let metricViewName: string;

  $: metricsView = useMetricsView($runtime.instanceId, metricViewName);
  $: displayName = $metricsView.data?.title;
</script>

<h1 class="ui-copy-dashboard-header">
  {displayName || metricViewName}
</h1>

<style lang="postcss">
  h1 {
    @apply overflow-hidden whitespace-nowrap text-ellipsis;
    @apply mr-2 mt-[-1px] leading-5;
  }
</style>
