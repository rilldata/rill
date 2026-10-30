<script lang="ts">
  import * as Popover from "@rilldata/web-common/components/popover";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { createRuntimeServiceGetExplore } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { timeAgo } from "@rilldata/web-common/lib/time/relative-time";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { Clock } from "lucide-svelte";

  export let dashboard: string;

  const runtimeClient = useRuntimeClient();

  $: lastRefreshedQuery = createRuntimeServiceGetExplore(
    runtimeClient,
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
  <!-- On phones the relative text costs a full header row, so it collapses to
       a clock icon; hover tooltips never fire on touch, so the icon opens a
       popover on tap instead. -->
  <Popover.Root>
    <Popover.Trigger
      class="sm:hidden p-1 -m-1 flex items-center text-fg-secondary"
      aria-label={m.dashboard_last_refreshed_ago({ time: timeAgo(data) })}
    >
      <Clock size="14px" />
    </Popover.Trigger>
    <Popover.Content class="flex w-fit flex-col gap-y-0.5 text-xs" padding="2">
      <span>{m.dashboard_last_refreshed_ago({ time: timeAgo(data) })}</span>
      <span class="text-fg-secondary">{data.toLocaleString()}</span>
    </Popover.Content>
  </Popover.Root>
  <Tooltip distance={8}>
    <div class="hidden sm:block text-[11px] text-fg-secondary">
      {m.dashboard_last_refreshed_ago({ time: timeAgo(data) })}
    </div>
    <TooltipContent slot="tooltip-content">
      {data.toLocaleString()}
    </TooltipContent>
  </Tooltip>
{/if}
