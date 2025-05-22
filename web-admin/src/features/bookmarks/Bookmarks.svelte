<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceRemoveBookmark,
    getAdminServiceListBookmarksQueryKey,
  } from "@rilldata/web-admin/client";
  import BookmarkDialog from "@rilldata/web-admin/features/bookmarks/BookmarkDialog.svelte";
  import BookmarksContent from "@rilldata/web-admin/features/bookmarks/BookmarksDropdownMenuContent.svelte";
  import { createHomeBookmarkModifier } from "@rilldata/web-admin/features/bookmarks/createOrUpdateHomeBookmark";
  import { getBookmarkDataForDashboard } from "@rilldata/web-admin/features/bookmarks/getBookmarkDataForDashboard";
  import HomeBookmarkButton from "@rilldata/web-admin/features/bookmarks/HomeBookmarkButton.svelte";
  import {
    type BookmarkEntry,
    categorizeBookmarks,
    getBookmarks,
  } from "@rilldata/web-admin/features/bookmarks/selectors";
  import {
    getProjectPermissions,
    useProjectId,
  } from "@rilldata/web-admin/features/projects/selectors";
  import { Button } from "@rilldata/web-common/components/button";
  import {
    DropdownMenu,
    DropdownMenuTrigger,
  } from "@rilldata/web-common/components/dropdown-menu";
  import { useMetricsViewTimeRange } from "@rilldata/web-common/features/dashboards/selectors";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { useExploreState } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { createUrlForExploreYAMLDefaultState } from "@rilldata/web-common/features/dashboards/stores/get-explore-state-from-yaml-config";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { createQueryServiceMetricsViewSchema } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { BookmarkIcon } from "lucide-svelte";

  export let metricsViewName: string;
  export let exploreName: string;

  let showDialog = false;
  let bookmark: BookmarkEntry | null = null;

  const { validSpecStore } = getStateManagers();

  $: organization = $page.params.organization;
  $: project = $page.params.project;
  $: ({ instanceId } = $runtime);

  $: exploreState = useExploreState(exploreName);
  $: projectId = useProjectId(organization, project);
  $: bookamrksResp = getBookmarks($projectId.data, exploreName);

  $: validExploreSpec = useExploreValidSpec(instanceId, exploreName);
  $: metricsViewSpec = $validExploreSpec.data?.metricsView ?? {};
  $: exploreSpec = $validExploreSpec.data?.explore ?? {};
  $: metricsViewTimeRange = useMetricsViewTimeRange(
    instanceId,
    metricsViewName,
    {},
    queryClient,
  );
  $: schemaResp = createQueryServiceMetricsViewSchema(
    instanceId,
    metricsViewName,
  );

  $: urlForExploreYAMLDefaultState = createUrlForExploreYAMLDefaultState(
    validExploreSpec,
    metricsViewTimeRange,
  );

  $: categorizedBookmarks = categorizeBookmarks(
    $bookamrksResp.data?.bookmarks ?? [],
    metricsViewSpec,
    exploreSpec,
    $schemaResp.data?.schema ?? {},
    $exploreState,
    $metricsViewTimeRange.data?.timeRangeSummary,
  );

  $: projectPermissions = getProjectPermissions(organization, project);
  $: manageProject = $projectPermissions.data?.manageProject;

  const queryClient = useQueryClient();
  $: homeBookmarkModifier = createHomeBookmarkModifier(exploreName);
  const bookmarkDeleter = createAdminServiceRemoveBookmark();

  async function createHomeBookmark() {
    await homeBookmarkModifier(
      getBookmarkDataForDashboard(
        $exploreState,
        $validSpecStore.data?.explore ?? {},
      ),
      $projectId.data,
      $bookamrksResp.data?.bookmarks ?? [],
    );
    eventBus.emit("notification", {
      message: "Home bookmark created",
    });
    return queryClient.refetchQueries({
      queryKey: getAdminServiceListBookmarksQueryKey({
        projectId: $projectId.data ?? "",
        resourceKind: ResourceKind.Explore,
        resourceName: exploreName,
      }),
    });
  }

  async function deleteBookmark(bookmark: BookmarkEntry) {
    // TODO: add confirmation?
    await $bookmarkDeleter.mutateAsync({
      bookmarkId: bookmark.resource.id,
    });
    eventBus.emit("notification", {
      message: `Bookmark ${bookmark.resource.displayName} deleted`,
    });
    return queryClient.refetchQueries({
      queryKey: getAdminServiceListBookmarksQueryKey({
        projectId: $projectId.data ?? "",
        resourceKind: ResourceKind.Explore,
        resourceName: exploreName,
      }),
    });
  }

  let open = false;
</script>

<HomeBookmarkButton
  {organization}
  {project}
  {exploreName}
  homeBookmark={categorizedBookmarks.home}
  urlForExploreYAMLDefaultState={$urlForExploreYAMLDefaultState ?? ""}
  onCreate={createHomeBookmark}
  onDelete={deleteBookmark}
  {manageProject}
/>

<DropdownMenu bind:open typeahead={false}>
  <DropdownMenuTrigger asChild let:builder>
    <Button
      builders={[builder]}
      compact
      type="secondary"
      label="Other bookmark dropdown"
      active={open}
    >
      <BookmarkIcon class="inline-flex" size="16px" />
    </Button>
  </DropdownMenuTrigger>
  <BookmarksContent
    onCreate={() => (showDialog = true)}
    onEdit={(editingBookmark) => {
      showDialog = true;
      bookmark = editingBookmark;
    }}
    onDelete={deleteBookmark}
    {categorizedBookmarks}
    {manageProject}
  />
</DropdownMenu>

{#if showDialog}
  <BookmarkDialog
    {bookmark}
    {metricsViewName}
    {exploreName}
    onClose={() => {
      showDialog = false;
      bookmark = null;
    }}
  />
{/if}
