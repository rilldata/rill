<script lang="ts">
  import Tag from "@rilldata/web-common/components/tag/Tag.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import ResourceTypeBadge from "@rilldata/web-common/features/entity-management/ResourceTypeBadge.svelte";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import ResourceListRow from "@rilldata/web-common/features/resources/ResourceListRow.svelte";
  import { timeAgo } from "@rilldata/web-common/lib/time/relative-time";

  export let name: string;
  export let title: string;
  export let lastRefreshed: string;
  export let description: string;
  export let error: string;
  export let isMetricsExplorer: boolean;
  export let isEmbedded: boolean;
  export let organization: string;
  export let project: string;

  $: lastRefreshedDate = lastRefreshed ? new Date(lastRefreshed) : null;

  $: dashboardSlug = isMetricsExplorer ? "explore" : "canvas";
  $: href = isEmbedded
    ? `/-/embed/${dashboardSlug}/${name}`
    : `/${organization}/${project}/${dashboardSlug}/${name}`;

  $: resourceKind = isMetricsExplorer
    ? ResourceKind.Explore
    : ResourceKind.Canvas;
</script>

<ResourceListRow
  {href}
  title={title !== "" ? title : name}
  errorMessage={error || undefined}
>
  {#snippet tags()}
    <ResourceTypeBadge kind={resourceKind} />
  {/snippet}

  {#snippet subtitle()}
    <span class="shrink-0">{name}</span>
    {#if lastRefreshedDate}
      <span class="shrink-0">•</span>
      <Tooltip distance={8}>
        <span class="shrink-0">Last refreshed {timeAgo(lastRefreshedDate)}</span
        >
        <TooltipContent slot="tooltip-content">
          {lastRefreshedDate.toLocaleString()}
        </TooltipContent>
      </Tooltip>
    {/if}
    {#if description}
      <span class="shrink-0">•</span>
      <span class="truncate">{description}</span>
    {/if}
  {/snippet}
</ResourceListRow>
