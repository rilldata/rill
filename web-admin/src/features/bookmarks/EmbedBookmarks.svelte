<script lang="ts">
  import { page } from "$app/state";
  import Bookmarks from "@rilldata/web-admin/features/bookmarks/Bookmarks.svelte";
  import { getEmbedBookmarksQueryOptions } from "@rilldata/web-admin/features/bookmarks/bookmark-client.ts";
  import {
    categorizeBookmarks,
    parseBookmarks,
  } from "@rilldata/web-admin/features/bookmarks/utils.ts";
  import { createRillDefaultExploreUrlParamsV2 } from "@rilldata/web-common/features/dashboards/url-state/get-rill-default-explore-url-params.ts";
  import { getOrgAndProjectForEmbed } from "@rilldata/web-common/features/embeds/selectors.ts";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { createQuery } from "@tanstack/svelte-query";
  import { readable } from "svelte/store";

  /**
   * Personal bookmarks for a dashboard in an embed.
   * Embeds don't have home or shared bookmarks, and all their bookmarks store url params, so no legacy proto data is handled.
   * The parent must key this component by the resource, since it is read once.
   */
  let {
    resource,
  }: {
    resource: {
      kind: ResourceKind.Explore | ResourceKind.Canvas;
      name: string;
    };
  } = $props();

  const runtimeClient = useRuntimeClient();

  // svelte-ignore state_referenced_locally
  const { kind: resourceKind, name: resourceName } = resource;
  const resourceNameStore = readable(resourceName);
  const isExplore = resourceKind === ResourceKind.Explore;

  // Used for the recently used bookmarks storage key.
  const orgAndProject = getOrgAndProjectForEmbed(runtimeClient);

  // Rill opinionated explore url params are removed from the url to keep it short.
  // They are added back when creating bookmarks, and removed from bookmarks when comparing them to the current url.
  // Canvas has no such defaults.
  const rillDefaultExploreUrlParams = isExplore
    ? createRillDefaultExploreUrlParamsV2(runtimeClient, resourceNameStore)
    : readable<URLSearchParams | undefined>(undefined);

  const bookmarksQuery = createQuery(
    getEmbedBookmarksQueryOptions(resourceKind, resourceNameStore),
  );

  let { organization, project } = $derived($orgAndProject);
  let bookmarks = $derived($bookmarksQuery.data?.bookmarks ?? []);
  let categorizedBookmarks = $derived(
    categorizeBookmarks(
      parseBookmarks(
        bookmarks,
        page.url.searchParams,
        $rillDefaultExploreUrlParams ?? null,
      ),
    ),
  );
</script>

<Bookmarks
  {organization}
  {project}
  projectId=""
  {resource}
  bookmarkData={{
    bookmarks,
    categorizedBookmarks,
    defaultUrlParams: $rillDefaultExploreUrlParams,
    showFiltersOnly: isExplore,
  }}
  showSharedBookmarks={false}
/>
