<script lang="ts">
  import DashboardIcon from "@rilldata/web-common/components/icons/DashboardIcon.svelte";
  import Tag from "@rilldata/web-common/components/tag/Tag.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { timeAgo } from "./utils";

  export let organization: string;
  export let project: string;
  export let name: string;
  export let title: string;
  export let lastRefreshed: string;
  export let description: string;
  export let error: string;

  $: lastRefreshedDate = new Date(lastRefreshed);
  $: isValidLastRefreshedDate = !isNaN(lastRefreshedDate.getTime());
</script>

<a
  href={`/${organization}/${project}/${name}`}
  class="flex flex-col gap-y-0.5 group px-4 py-[5px]"
>
  <div class="flex gap-x-2 items-center">
    <DashboardIcon size={"14px"} className="text-slate-500" />
    <div
      class="text-gray-700 text-sm font-semibold group-hover:text-primary-600"
    >
      {title !== "" ? title : name}
    </div>
    {#if error !== ""}
      <Tag color="red">Error</Tag>
    {/if}
  </div>
  <div class="pl-[22px] flex gap-x-1 text-gray-500 text-xs font-normal">
    <span class="shrink-0">{name}</span>
    {#if isValidLastRefreshedDate}
      <span>•</span>
      <Tooltip distance={8}>
        <span class="shrink-0">Last refreshed {timeAgo(lastRefreshedDate)}</span
        >
        <TooltipContent slot="tooltip-content">
          {lastRefreshedDate.toLocaleString()}
        </TooltipContent>
      </Tooltip>
    {/if}
    {#if description}
      <span>•</span>
      <span class="line-clamp-1">{description}</span>
    {/if}
  </div>
</a>
