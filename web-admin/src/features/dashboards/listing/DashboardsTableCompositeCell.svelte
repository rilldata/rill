<script lang="ts">
  import Tag from "@rilldata/web-common/components/tag/Tag.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import ResourceTypeBadge from "@rilldata/web-common/features/entity-management/ResourceTypeBadge.svelte";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { timeAgo } from "@rilldata/web-common/lib/time/relative-time";
  import { Star } from "lucide-svelte";
  import { ArrayRuneStore } from "web-common/src/lib/store-utils/types.svelte.ts";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import type { RecentlyUsedDashboards } from "./dashboard-favourites.ts";

  export let name: string;
  export let title: string;
  export let lastRefreshed: string;
  export let description: string;
  export let error: string;
  export let isMetricsExplorer: boolean;
  export let isEmbedded: boolean;
  export let organization: string;
  export let project: string;
  export let tags: string[] = [];
  export let dashboardFavourites: ArrayRuneStore<string> | undefined =
    undefined;
  export let recentlyUsedDashboards: RecentlyUsedDashboards | undefined =
    undefined;

  $: lastRefreshedDate = lastRefreshed ? new Date(lastRefreshed) : null;

  $: dashboardSlug = isMetricsExplorer ? "explore" : "canvas";
  $: href = isEmbedded
    ? `/-/embed/${dashboardSlug}/${name}`
    : `/${organization}/${project}/${dashboardSlug}/${name}`;

  $: resourceKind = isMetricsExplorer
    ? ResourceKind.Explore
    : ResourceKind.Canvas;

  $: favourites = dashboardFavourites?.value ?? [];
  $: isFavourite = favourites.includes(name?.toLowerCase());

  $: lastUsed =
    recentlyUsedDashboards?.recentlyUsed?.value?.[name.toLowerCase()];
  $: lastUsedDate = lastUsed ? new Date(lastUsed) : null;

  let hovered = false;

  function toggleFavourite() {
    dashboardFavourites?.toggle(name?.toLowerCase());
  }

  function onDashboardFavouriteToggle(e: MouseEvent) {
    e.stopPropagation();
    e.preventDefault();
    toggleFavourite();
  }

  function onDashboardFavouriteKeydown(e: KeyboardEvent) {
    if (e.key !== "Enter" && e.key !== " ") return;
    e.stopPropagation();
    e.preventDefault();
    toggleFavourite();
  }
</script>

<!-- TODO: Support non-hover actions like using keyboard for showing favourite icon -->
<a
  class="flex flex-col gap-y-1 group px-4 py-2.5 w-full h-full"
  {href}
  onmouseenter={() => (hovered = true)}
  onmouseleave={() => (hovered = false)}
>
  <div class="flex gap-x-2 items-center min-h-[20px]">
    <ResourceTypeBadge kind={resourceKind} />
    <span
      class="text-fg-secondary text-sm font-semibold group-hover:text-accent-primary-action truncate"
    >
      {title !== "" ? title : name}
    </span>
    {#if error !== ""}
      <Tag color="red">{m.dashboard_error_tag()}</Tag>
    {/if}
    {#each tags as tag (tag)}
      <Tag color="gray">{tag}</Tag>
    {/each}
    <div class="grow"></div>
    {#if dashboardFavourites && (hovered || isFavourite)}
      <span
        role="button"
        tabindex="0"
        aria-label={isFavourite
          ? m.dashboard_remove_favourite()
          : m.dashboard_add_favourite()}
        aria-pressed={isFavourite}
        class="shrink-0 flex cursor-pointer"
        onclick={onDashboardFavouriteToggle}
        onkeydown={onDashboardFavouriteKeydown}
      >
        <Star
          size="16px"
          class={isFavourite ? "fill-accent-primary-action" : "none"}
        />
      </span>
    {/if}
  </div>
  <div
    class="flex gap-x-1 text-fg-tertiary text-xs font-normal min-h-[16px] overflow-hidden"
  >
    <span class="shrink-0">{name}</span>
    {#if lastRefreshedDate}
      <span class="shrink-0">•</span>
      <Tooltip distance={8}>
        <span class="shrink-0"
          >{m.dashboard_last_refreshed_ago({
            time: timeAgo(lastRefreshedDate),
          })}</span
        >
        <TooltipContent slot="tooltip-content">
          {lastRefreshedDate.toLocaleString()}
        </TooltipContent>
      </Tooltip>
    {/if}
    {#if lastUsedDate}
      <span class="shrink-0">•</span>
      <Tooltip distance={8}>
        <span class="shrink-0"
          >{m.dashboard_last_used_ago({
            time: timeAgo(lastUsedDate),
          })}</span
        >
        <TooltipContent slot="tooltip-content">
          {lastUsedDate.toLocaleString()}
        </TooltipContent>
      </Tooltip>
    {/if}
    {#if description}
      <span class="shrink-0">•</span>
      <span class="truncate">{description}</span>
    {/if}
  </div>
</a>
