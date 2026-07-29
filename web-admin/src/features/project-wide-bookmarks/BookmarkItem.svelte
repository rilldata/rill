<script lang="ts">
  import * as Dropdown from "@rilldata/web-common/components/dropdown-menu";
  import type { V1Bookmark } from "@rilldata/web-admin/client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { ParsedBookmark } from "@rilldata/web-admin/features/project-wide-bookmarks/ParsedBookmark.svelte.ts";
  import HomeBookmark from "@rilldata/web-common/components/icons/HomeBookmark.svelte";
  import FilterFilled from "@rilldata/web-common/components/icons/FilterFilled.svelte";
  import FilterOutline from "@rilldata/web-common/components/icons/FilterOutline.svelte";
  import BookmarkFilled from "@rilldata/web-common/components/icons/BookmarkFilled.svelte";
  import BookmarkOutline from "@rilldata/web-common/components/icons/BookmarkOutline.svelte";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import Trash from "@rilldata/web-common/components/icons/Trash.svelte";
  import FeatherEditIcon from "@rilldata/web-common/components/icons/FeatherEditIcon.svelte";
  import { ReplaceIcon } from "lucide-svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import { onDestroy } from "svelte";

  let {
    bookmark,
    onEdit,
    onReplace,
    onDelete,
  }: {
    bookmark: V1Bookmark;
    onEdit: (bookmark: V1Bookmark) => void;
    onReplace: (parsedBookmark: ParsedBookmark) => void;
    onDelete: (bookmark: V1Bookmark) => void;
  } = $props();

  const runtimeClient = useRuntimeClient();
  let parsedBookmark = $derived(new ParsedBookmark(runtimeClient, bookmark));

  let description = $derived(
    bookmark.description || `${bookmark.resourceName}`,
  );

  const IconsByType = {
    home: {
      active: HomeBookmark,
      inactive: HomeBookmark,
    },
    filter: {
      active: FilterFilled,
      inactive: FilterOutline,
    },
    complete: {
      active: BookmarkFilled,
      inactive: BookmarkOutline,
    },
  };
  let icons = $derived(
    bookmark.default
      ? IconsByType.home
      : parsedBookmark.filtersOnly
        ? IconsByType.filter
        : IconsByType.complete,
  );
  let IconComponent = $derived(
    parsedBookmark.isActive ? icons.active : icons.inactive,
  );

  let hovering = $state(false);
  let menuOpen = $state(false);

  onDestroy(() => {
    parsedBookmark.cleanup?.();
  });
</script>

<Dropdown.Item
  class="flex justify-between gap-x-2 w-full"
  onmouseenter={() => (hovering = true)}
  onmouseleave={() => (hovering = false)}
>
  <a href={parsedBookmark.fullUrl} class="flex flex-row gap-x-2 w-full">
    <IconComponent size="16px" className="text-fg-primary" />
    <div class="flex flex-col gap-y-0.5">
      <div
        class="text-xs font-medium text-fg-primary h-4 text-ellipsis overflow-hidden"
      >
        {bookmark.displayName}
      </div>
      {#if description}
        <div
          class="text-[11px] font-normal text-fg-secondary h-4 text-ellipsis overflow-hidden"
        >
          {description}
        </div>
      {/if}
    </div>
  </a>
  <div class="flex flex-row gap-x-2 min-w-8 min-h-full">
    {#if hovering || menuOpen}
      <Dropdown.Root bind:open={menuOpen}>
        <Dropdown.Trigger>
          {#snippet child({ props })}
            <Button
              {...props}
              onClick={(e) => {
                e.stopPropagation();
              }}
              compact
              small
            >
              <ThreeDot size="14px" />
            </Button>
          {/snippet}
        </Dropdown.Trigger>
        <Dropdown.Content class="text-xs" align="end" side="bottom">
          <Dropdown.Item onclick={() => onEdit(bookmark)}>
            <FeatherEditIcon size="12px" />
            Edit
          </Dropdown.Item>
          <Dropdown.Item onclick={() => onReplace(parsedBookmark)}>
            <ReplaceIcon size="12px" />
            Replace
          </Dropdown.Item>
          <Dropdown.Item
            class="text-destructive"
            onclick={() => onDelete(bookmark)}
          >
            <Trash size="12px" />
            Delete
          </Dropdown.Item>
        </Dropdown.Content>
      </Dropdown.Root>
    {/if}
  </div>
</Dropdown.Item>
