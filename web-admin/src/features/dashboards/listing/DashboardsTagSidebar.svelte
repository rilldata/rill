<script lang="ts">
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import DashboardsTagRow from "./DashboardsTagRow.svelte";
  import { UrlParamsState } from "web-common/src/lib/store-utils/url-params-state.svelte.ts";
  import { getAllTagsForResources } from "@rilldata/web-common/features/resources/resource-tag-utils.ts";
  import {
    getDashboardTagFavouritesStore,
    sortByFavourites,
  } from "./dashboard-favourites.ts";
  import { page } from "$app/state";
  import { flip } from "svelte/animate";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  let {
    resources,
    searchText,
  }: {
    resources: V1Resource[];
    searchText: string;
  } = $props();

  const selectedTagsState = UrlParamsState.createStringArrayParam("tags");

  // Derive the tag rows (name + count) from the dashboards' meta tags. The tag
  // index dedupes and counts; we sort alphabetically for a stable listing.
  let tags = $derived(getAllTagsForResources(resources));

  let filteredTags = $derived(
    searchText.trim()
      ? tags.filter((t) =>
          t.name.toLowerCase().includes(searchText.trim().toLowerCase()),
        )
      : tags,
  );

  let { organization, project } = $derived(page.params);
  let tagsFavourites = $derived(
    getDashboardTagFavouritesStore(organization, project),
  );

  let sortedTags = $derived(
    sortByFavourites(filteredTags, tagsFavourites.value, (t) => t.name),
  );
</script>

<div class="tags-scroll">
  <h3 class="column-header">{m.dashboard_tags()}</h3>

  {#if sortedTags.length === 0}
    <p class="text-fg-secondary my-1 px-2 text-xs">
      {m.dashboard_no_matching_tags()}
    </p>
  {:else}
    {#each sortedTags as tag (tag.name)}
      <div animate:flip={{ duration: 200 }}>
        <DashboardsTagRow
          name={tag.name}
          count={tag.totalCount}
          selected={selectedTagsState.value.includes(tag.name)}
          isFavourite={tagsFavourites.value.includes(tag.name)}
          onToggle={() => selectedTagsState.toggle(tag.name)}
          onFavouriteToggle={() => tagsFavourites.toggle(tag.name)}
        />
      </div>
    {/each}
  {/if}
</div>

<style lang="postcss">
  .tags-scroll {
    @apply flex flex-col h-full w-full py-2 px-2 gap-y-0.5;
    @apply bg-surface-background overflow-y-auto overflow-x-hidden;
  }

  .column-header {
    @apply uppercase font-semibold text-[10px];
    @apply px-1.5 pt-1 pb-1 text-fg-secondary;
  }
</style>
