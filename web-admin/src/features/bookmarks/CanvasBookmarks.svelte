<script lang="ts">
  import Bookmarks from "@rilldata/web-admin/features/bookmarks/Bookmarks.svelte";
  import { getCanvasCategorisedBookmarks } from "@rilldata/web-admin/features/bookmarks/selectors.ts";
  import { getProjectIdQueryOptions } from "@rilldata/web-admin/features/projects/selectors.ts";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
  import { createQuery } from "@tanstack/svelte-query";
  import { writable } from "svelte/store";

  export let organization: string;
  export let project: string;
  export let canvasName: string;

  const orgAndProjectNameStore = writable({ organization, project });
  $: orgAndProjectNameStore.set({ organization, project });

  const projectIdQuery = createQuery(
    getProjectIdQueryOptions(orgAndProjectNameStore),
  );
  $: projectId = $projectIdQuery.data ?? "";

  const canvasNameStore = writable(canvasName);
  $: canvasNameStore.set(canvasName);

  const categorizedBookmarksStore = getCanvasCategorisedBookmarks(
    orgAndProjectNameStore,
    canvasNameStore,
  );

  $: ({
    data: { bookmarks, categorizedBookmarks },
  } = $categorizedBookmarksStore);
</script>

<Bookmarks
  {organization}
  {project}
  {projectId}
  resource={{ name: canvasName, kind: ResourceKind.Canvas }}
  bookmarkData={{
    bookmarks,
    categorizedBookmarks,
    showFiltersOnly: false,
    defaultHomeBookmarkUrl: "?default=true",
  }}
/>
