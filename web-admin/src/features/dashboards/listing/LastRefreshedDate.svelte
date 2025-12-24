<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { createRuntimeServiceGetExplore } from "@rilldata/web-common/runtime-client";
  import httpClient from "@rilldata/web-common/runtime-client/http-client";
  import { timeAgo } from "./utils";

  export let dashboard: string;

  const instanceId = httpClient.getInstanceId();

  $: lastRefreshedQuery = createRuntimeServiceGetExplore(
    instanceId,
    { name: dashboard },
    {
      query: {
        select: (data) => {
          const refreshDate =
            data?.metricsView?.metricsView?.state?.dataRefreshedOn;
          return refreshDate ? new Date(refreshDate) : null;
        },
      },
    },
  );
  $: ({ data } = $lastRefreshedQuery);
</script>

{#if data}
  <Tooltip distance={8}>
    <div class="text-[11px] text-gray-600">
      Last refreshed {timeAgo(data)}
    </div>
    <TooltipContent slot="tooltip-content">
      {data.toLocaleString()}
    </TooltipContent>
  </Tooltip>
{/if}
