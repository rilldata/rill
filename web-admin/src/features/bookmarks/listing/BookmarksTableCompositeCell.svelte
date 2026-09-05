<script lang="ts">
  import type { BookmarkListRow } from "@rilldata/web-admin/features/bookmarks/listing/bookmark-listing-utils.ts";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import BookmarkOutline from "@rilldata/web-common/components/icons/BookmarkOutline.svelte";
  import FilterOutline from "@rilldata/web-common/components/icons/FilterOutline.svelte";
  import HomeBookmark from "@rilldata/web-common/components/icons/HomeBookmark.svelte";
  import Trash from "@rilldata/web-common/components/icons/Trash.svelte";
  import Tag from "@rilldata/web-common/components/tag/Tag.svelte";
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2";
  import ResourceTypeBadge from "@rilldata/web-common/features/entity-management/ResourceTypeBadge.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { timeAgo } from "@rilldata/web-common/lib/time/relative-time";
  import { Pencil } from "lucide-svelte";

  let {
    row,
    onOpen,
    onEdit,
    onDelete,
  }: {
    row: BookmarkListRow;
    onOpen: (row: BookmarkListRow) => void;
    onEdit: (row: BookmarkListRow) => void;
    onDelete: (row: BookmarkListRow) => void;
  } = $props();

  let Icon = $derived(
    row.category === "home"
      ? HomeBookmark
      : row.filtersOnly
        ? FilterOutline
        : BookmarkOutline,
  );
  let displayName = $derived(row.bookmark.displayName ?? "");
  let updatedOn = $derived(
    row.bookmark.updatedOn ? new Date(row.bookmark.updatedOn) : null,
  );
  let lastUsed = $derived(row.lastUsed ? new Date(row.lastUsed) : null);

  let hovered = $state(false);
</script>

<div
  class="flex flex-row items-center gap-x-2 group px-4 w-full h-full"
  onmouseenter={() => (hovered = true)}
  onmouseleave={() => (hovered = false)}
  role="row"
  tabindex="-1"
>
  <a
    class="flex flex-row items-center gap-x-3 min-w-0 grow h-full py-2.5"
    href={row.href}
    aria-label={m.bookmark_entry_aria_label({ name: displayName })}
    onclick={() => onOpen(row)}
  >
    <Icon size="16px" className="shrink-0 text-fg-secondary" />
    <div class="flex flex-col gap-y-1 min-w-0 grow">
      <div class="flex gap-x-2 items-center min-h-[20px]">
        <span
          class="text-fg-secondary text-sm font-semibold group-hover:text-accent-primary-action truncate"
        >
          {displayName}
        </span>
        {#if row.category === "home"}
          <Tag color="blue">{m.bookmark_tag_home()}</Tag>
        {:else if row.category === "managed"}
          <Tag color="gray">{m.bookmark_tag_managed()}</Tag>
        {/if}
        {#if row.isLegacy}
          <Tooltip.Root>
            <Tooltip.Trigger>
              {#snippet child({ props })}
                <span {...props}>
                  <Tag color="amber">{m.bookmark_tag_legacy()}</Tag>
                </span>
              {/snippet}
            </Tooltip.Trigger>
            <Tooltip.Content side="bottom">
              {m.bookmark_legacy_tooltip()}
            </Tooltip.Content>
          </Tooltip.Root>
        {/if}
      </div>
      <div
        class="flex gap-x-1 items-center text-fg-tertiary text-xs font-normal min-h-[16px] overflow-hidden"
      >
        {#if row.dashboardKind}
          <ResourceTypeBadge kind={row.dashboardKind} />
        {/if}
        <span class="shrink-0 truncate max-w-64">{row.dashboardTitle}</span>
        {#if updatedOn}
          <span class="shrink-0">•</span>
          <Tooltip.Root>
            <Tooltip.Trigger>
              {#snippet child({ props })}
                <span {...props} class="shrink-0">
                  {m.bookmark_updated_ago({ time: timeAgo(updatedOn) })}
                </span>
              {/snippet}
            </Tooltip.Trigger>
            <Tooltip.Content side="bottom">
              {updatedOn.toLocaleString()}
            </Tooltip.Content>
          </Tooltip.Root>
        {/if}
        {#if lastUsed}
          <span class="shrink-0">•</span>
          <Tooltip.Root>
            <Tooltip.Trigger>
              {#snippet child({ props })}
                <span {...props} class="shrink-0">
                  {m.bookmark_last_used_ago({ time: timeAgo(lastUsed) })}
                </span>
              {/snippet}
            </Tooltip.Trigger>
            <Tooltip.Content side="bottom">
              {lastUsed.toLocaleString()}
            </Tooltip.Content>
          </Tooltip.Root>
        {/if}
        {#if row.bookmark.description}
          <span class="shrink-0">•</span>
          <span class="truncate">{row.bookmark.description}</span>
        {/if}
      </div>
    </div>
  </a>
  {#if row.canManage}
    <div class="flex flex-row justify-end items-center gap-x-1 w-16 shrink-0">
      {#if hovered}
        <Button
          square
          type="tertiary"
          label={m.bookmark_edit()}
          onClick={() => onEdit(row)}
        >
          <Pencil size="16px" />
        </Button>
        <Button
          square
          type="tertiary"
          label={m.bookmark_delete_bookmark()}
          onClick={() => onDelete(row)}
        >
          <Trash size="16px" />
        </Button>
      {/if}
    </div>
  {/if}
</div>
