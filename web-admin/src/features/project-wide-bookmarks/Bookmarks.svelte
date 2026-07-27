<script lang="ts">
  import * as Dropdown from "@rilldata/web-common/components/dropdown-menu";
  import { Button } from "@rilldata/web-common/components/button";
  import { BookmarkIcon } from "lucide-svelte";
  import {
    createAdminServiceGetProject,
    createAdminServiceListBookmarksInfinite,
  } from "@rilldata/web-admin/client";
  import BookmarkItem from "@rilldata/web-admin/features/project-wide-bookmarks/BookmarkItem.svelte";

  let { organization, project }: { organization: string; project: string } =
    $props();

  let open = $state(false);
  const BookmarksPageSize = 1000;

  let projectQuery = $derived(
    createAdminServiceGetProject(organization, project),
  );

  let bookmarksQuery = $derived(
    createAdminServiceListBookmarksInfinite(
      {
        projectId: $projectQuery.data?.project?.id,
        pageSize: BookmarksPageSize,
      },
      {
        query: {
          enabled: Boolean($projectQuery.data?.project?.id),
          getNextPageParam: (lastPage) => {
            if (lastPage.nextPageToken !== "") {
              return lastPage.nextPageToken;
            }
            return undefined;
          },
        },
      },
    ),
  );
  let allBookmarks = $derived(
    $bookmarksQuery.data?.pages.flatMap((page) => page.bookmarks ?? []) ?? [],
  );
</script>

<Dropdown.Root bind:open>
  <Dropdown.Trigger>
    {#snippet child({ props })}
      <Button
        {...props}
        compact
        square
        type="secondary"
        label="Other bookmark dropdown"
        active={open}
      >
        <BookmarkIcon class="flex-none" size="16px" />
      </Button>
    {/snippet}
  </Dropdown.Trigger>
  <Dropdown.Content class="gap-2">
    {#each allBookmarks as bookmark (bookmark.id)}
      <BookmarkItem {bookmark} />
    {/each}
  </Dropdown.Content>
</Dropdown.Root>
