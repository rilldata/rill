<script lang="ts">
  import DashboardIcon from "@rilldata/web-common/components/icons/DashboardIcon.svelte";
  import Tag from "@rilldata/web-common/components/tag/Tag.svelte";
  import { timeAgo } from "./utils";

  export let organization: string;
  export let project: string;
  export let name: string;
  export let title: string;
  export let lastRefreshed: number;
  export let description: string;
  export let error: string;
</script>

<div class="flex flex-col gap-y-0.5">
  <div class="flex gap-x-2 items-center">
    <DashboardIcon size={"14px"} className="text-slate-500" />
    <a
      href={`/${organization}/${project}/${name}`}
      class="text-gray-700 text-sm font-semibold hover:text-blue-600"
    >
      {title}
    </a>
    {#if error !== ""}
      <Tag color="red">Error</Tag>
    {/if}
  </div>
  <div class="pl-[22px] flex gap-x-1 text-gray-500 text-xs font-normal">
    <span class="shrink-0">{name}</span>
    <span>•</span>
    <span class="shrink-0">Last refreshed {timeAgo(lastRefreshed)}</span>
    {#if description}
      <span>•</span>
      <span class="line-clamp-1">{description}</span>
    {/if}
  </div>
</div>
