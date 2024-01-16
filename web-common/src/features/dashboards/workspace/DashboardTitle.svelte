<script lang="ts">
  import { runtime } from "../../../runtime-client/runtime-store";
  import { useMetaQuery } from "../selectors";

  export let metricViewName: string;

  $: metaQuery = useMetaQuery($runtime.instanceId, metricViewName);
  $: displayName = $metaQuery.data?.title;
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
