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
  import BookmarkItem from "@rilldata/web-admin/features/bookmarks/BookmarksDropdownMenuItem.svelte";
  import {
    getBookmarks,
    searchBookmarks,
  } from "@rilldata/web-admin/features/bookmarks/selectors";
  import { Search } from "@rilldata/web-common/components/search";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { BookmarkPlusIcon } from "lucide-svelte";
  import { createEventDispatcher } from "svelte";
  import HomeBookmarkPlus from "@rilldata/web-common/components/icons/HomeBookmarkPlus.svelte";

  const dispatch = createEventDispatcher();
  const queryClient = useQueryClient();

  $: organization = $page.params.organization;
  $: project = $page.params.project;

  let searchText: string;
  let bookmarks: ReturnType<typeof getBookmarks>;
  $: bookmarks = getBookmarks(
    queryClient,
    $runtime?.instanceId,
    organization,
    project,
    $page.params.dashboard,
  );
  $: filteredBookmarks = searchBookmarks($bookmarks.data, searchText);

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
  <DropdownMenuSeparator />
  <div class="p-2">
    <Search
      autofocus={false}
      bind:value={searchText}
      showBorderOnFocus={false}
    />
  </div>
  {#if filteredBookmarks}
    <DropdownMenuSeparator />
    <DropdownMenuGroup>
      <DropdownMenuLabel class="text-gray-500 text-[10px] h-6 uppercase">
        Your bookmarks
      </DropdownMenuLabel>
      {#if filteredBookmarks.personal?.length}
        {#each filteredBookmarks.personal as bookmark}
          {#key bookmark.resource.id}
            <BookmarkItem {bookmark} on:edit on:select on:delete />
          {/key}
        {/each}
      {:else}
        <div class="my-2 ui-copy-disabled text-center">
          You have no bookmarks for this dashboard
        </div>
      {/if}
    </DropdownMenuGroup>
    <DropdownMenuSeparator />
    <DropdownMenuGroup>
      <DropdownMenuLabel class="text-gray-500">
        <div class="text-[10px] h-4 uppercase">Managed bookmarks</div>
        <div class="text-[11px] font-normal">Created by project admin</div>
      </DropdownMenuLabel>
      {#if filteredBookmarks.shared?.length || filteredBookmarks.home}
        {#if filteredBookmarks.home}
          {#key filteredBookmarks.home.resource.id}
            <BookmarkItem
              bookmark={filteredBookmarks.home}
              on:edit
              on:select
              on:delete
              readOnly={!manageProject}
            />
          {/key}
        {/if}
        {#each filteredBookmarks.shared as bookmark}
          {#key bookmark.resource.id}
            <BookmarkItem
              {bookmark}
              on:edit
              on:select
              on:delete
              readOnly={!manageProject}
            />
          {/key}
        {/each}
      {:else}
        <div class="my-2 ui-copy-disabled text-center">
          There are no shared bookmarks for this dashboard.
        </div>
      {/if}
    </DropdownMenuGroup>
  {/if}
</DropdownMenuContent>
