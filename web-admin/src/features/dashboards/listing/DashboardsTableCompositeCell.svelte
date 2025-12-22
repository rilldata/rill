<script lang="ts">
  import { page } from "$app/stores";
  import Tag from "@rilldata/web-common/components/tag/Tag.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import ResourceTypeBadge from "@rilldata/web-common/features/entity-management/ResourceTypeBadge.svelte";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { timeAgo } from "./utils";

  export let name: string;
  export let title: string;
  export let lastRefreshed: string;
  export let description: string;
  export let error: string;
  export let isMetricsExplorer: boolean;
  export let isEmbedded: boolean;

  $: organization = $page.params.organization;
  $: project = $page.params.project;

  $: lastRefreshedDate = lastRefreshed ? new Date(lastRefreshed) : null;

  $: dashboardSlug = isMetricsExplorer ? "explore" : "canvas";
  $: href = isEmbedded
    ? `/-/embed/${dashboardSlug}/${name}`
    : `/${organization}/${project}/${dashboardSlug}/${name}`;

  $: resourceKind = isMetricsExplorer
    ? ResourceKind.Explore
    : ResourceKind.Canvas;
</script>

<a class="flex flex-col gap-y-1 group px-4 py-2.5 w-full h-full" {href}>
  <div class="flex gap-x-2 items-center min-h-[20px]">
    <ResourceTypeBadge kind={resourceKind} />
    <span
      class="text-gray-700 text-sm font-semibold group-hover:text-primary-600 truncate"
    >
      {title !== "" ? title : name}
    </span>
    {#if error !== ""}
      <Tag color="red">Error</Tag>
    {/if}
  </div>
  <div
    class="flex gap-x-1 text-gray-500 text-xs font-normal min-h-[16px] overflow-hidden"
  >
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
  </div>
</a>
