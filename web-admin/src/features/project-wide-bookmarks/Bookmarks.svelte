<script lang="ts">
  import * as Dropdown from "@rilldata/web-common/components/dropdown-menu";
  import { getBookmarksInfiniteQueryOptions } from "@rilldata/web-admin/features/project-wide-bookmarks/selectors.ts";
  import { writable } from "svelte/store";
  import { Button } from "@rilldata/web-common/components/button";
  import { BookmarkIcon } from "lucide-svelte";
  import { createInfiniteQuery } from "@tanstack/svelte-query";

  export let organization: string;
  export let project: string;

  let open = false;

  const orgAndProjectNameStore = writable({ organization: "", project: "" });
  $: orgAndProjectNameStore.set({ organization, project });

  const bookmarksQuery = createInfiniteQuery(
    getBookmarksInfiniteQueryOptions(orgAndProjectNameStore),
  );
  $: allBookmarks =
    $bookmarksQuery.data?.pages.flatMap((page) => page.bookmarks ?? []) ?? [];
</script>

<Dropdown.Root bind:open>
  <Dropdown.Trigger asChild let:builder>
    <Button
      builders={[builder]}
      compact
      square
      type="secondary"
      label="Other bookmark dropdown"
      active={open}
    >
      <BookmarkIcon class="flex-none" size="16px" />
    </Button>
  </Dropdown.Trigger>
  <Dropdown.Content class="gap-2">
    {#each allBookmarks as bookmark (bookmark.id)}
      <Dropdown.Item>
        <div>{bookmark.displayName}</div>
        <div>{bookmark.description}</div>
      </Dropdown.Item>
    {/each}
  </Dropdown.Content>
</Dropdown.Root>
