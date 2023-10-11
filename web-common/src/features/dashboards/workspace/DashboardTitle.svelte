<script lang="ts">
  import { runtime } from "../../../runtime-client/runtime-store";
  import { useMetaQuery } from "../selectors";

  export let metricViewName: string;

  $: metaQuery = useMetaQuery($runtime.instanceId, metricViewName);
  $: displayName = $metaQuery.data?.title;
</script>

<h1 style:line-height="1.1" style:margin-top="-1px">
  <div class="ui-copy-dashboard-header">
    {displayName || metricViewName}
  </div>
</h1>
