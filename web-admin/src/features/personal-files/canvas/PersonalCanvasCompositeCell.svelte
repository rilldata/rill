<script lang="ts">
  import * as m from "@rilldata/web-common/paraglide/messages.js";
  import Tag from "@rilldata/web-common/components/tag/Tag.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import ResourceTypeBadge from "@rilldata/web-common/features/entity-management/ResourceTypeBadge.svelte";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { timeAgo } from "@rilldata/web-common/lib/time/relative-time";

  let {
    name,
    title,
    lastRefreshed,
    error,
    isMetricsExplorer,
    organization,
    project,
  }: {
    name: string;
    title: string;
    lastRefreshed: string;
    error?: string;
    isMetricsExplorer: boolean;
    organization: string;
    project: string;
  } = $props();

  let lastRefreshedDate = $derived(
    lastRefreshed ? new Date(lastRefreshed) : null,
  );

  let href = $derived(`/${organization}/${project}/-/personal/${name}`);

  let resourceKind = $derived(
    isMetricsExplorer ? ResourceKind.Explore : ResourceKind.Canvas,
  );
</script>

<a class="flex flex-col gap-y-1 group px-4 py-2.5 w-full h-full" {href}>
  <div class="flex gap-x-2 items-center min-h-[20px]">
    <ResourceTypeBadge kind={resourceKind} />
    <span
      class="text-fg-secondary text-sm font-semibold group-hover:text-accent-primary-action truncate"
    >
      {title !== "" ? title : name}
    </span>
    {#if error}
      <Tag color="red">Error</Tag>
    {/if}
  </div>
  <div
    class="flex gap-x-1 text-fg-tertiary text-xs font-normal min-h-[16px] overflow-hidden"
  >
    {#if lastRefreshedDate}
      <Tooltip distance={8}>
        <span class="shrink-0">
          {m.dashboard_last_refreshed_ago({ time: timeAgo(lastRefreshedDate) })}
        </span>
        <TooltipContent slot="tooltip-content">
          {lastRefreshedDate.toLocaleString()}
        </TooltipContent>
      </Tooltip>
    {/if}
  </div>
</a>
