<script lang="ts">
  import { page } from "$app/stores";
  import { getProjectPermissions } from "@rilldata/web-admin/features/projects/selectors";
  import {
    DropdownMenuContent,
    DropdownMenuGroup,
    DropdownMenuItem,
    DropdownMenuLabel,
    DropdownMenuSeparator,
  } from "@rilldata/web-common/components/dropdown-menu";
  import BookmarkItem from "@rilldata/web-admin/features/bookmarks/BookmarkItem.svelte";
  import { getBookmarks } from "@rilldata/web-admin/features/bookmarks/selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { BookmarkPlusIcon } from "lucide-svelte";
  import { createEventDispatcher } from "svelte";
  import HomeBookmarkPlus from "@rilldata/web-common/components/icons/HomeBookmarkPlus.svelte";

  const dispatch = createEventDispatcher();
  const queryClient = useQueryClient();

  $: organization = $page.params.organization;
  $: project = $page.params.project;

  let bookmarks: ReturnType<typeof getBookmarks>;
  $: bookmarks = getBookmarks(
    queryClient,
    $runtime?.instanceId,
    organization,
    project,
    $page.params.dashboard,
  );

  $: projectPermissions = getProjectPermissions(organization, project);
  $: manageProject = $projectPermissions.data?.manageProject;
</script>

<DropdownMenuContent class="w-[450px]">
  <DropdownMenuItem on:click={() => dispatch("create")}>
    <div class="flex flex-row gap-x-2 items-center">
      <BookmarkPlusIcon size="16px" strokeWidth={1.5} />
      <div class="text-xs">Bookmark current view</div>
    </div>
  </DropdownMenuItem>
  {#if manageProject}
    <DropdownMenuItem
      on:click={() => dispatch("create-home")}
      slot="manage-project"
    >
      <div class="flex flex-row gap-x-2 items-center">
        <HomeBookmarkPlus size="16px" />
        <div class="text-xs">Bookmark current view as Home</div>
      </div>
    </DropdownMenuItem>
  {/if}
  {#if $bookmarks.data}
    {#if $bookmarks.data.personal?.length}
      <DropdownMenuSeparator />
      <DropdownMenuGroup>
        <DropdownMenuLabel class="capitalize text-gray-500 text-sm">
          Your bookmarks
        </DropdownMenuLabel>
        {#each $bookmarks.data.personal as bookmark}
          <BookmarkItem {bookmark} on:edit on:select on:delete />
        {/each}
      </DropdownMenuGroup>
    {/if}
    {#if $bookmarks.data.shared?.length || $bookmarks.data.home}
      <DropdownMenuSeparator />
      <DropdownMenuGroup>
        <DropdownMenuLabel class="capitalize text-gray-500 ">
          <div class="text-sm capitalize font-semibold">Default bookmarks</div>
          <div class="text-[11px] font-normal">Created by project admin</div>
        </DropdownMenuLabel>
        {#if $bookmarks.data.home}
          <BookmarkItem
            bookmark={$bookmarks.data.home}
            on:edit
            on:select
            on:delete
            readonly={!manageProject}
          />
        {/if}
        {#each $bookmarks.data.shared as bookmark}
          <BookmarkItem
            {bookmark}
            on:edit
            on:select
            on:delete
            readonly={!manageProject}
          />
        {/each}
      </DropdownMenuGroup>
    {/if}
  {/if}
</DropdownMenuContent>
