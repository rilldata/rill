<script lang="ts">
  import { Search } from "@rilldata/web-common/components/search";
  import { buildTagIndex } from "@rilldata/web-common/components/menu/tag-utils";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import DashboardsTagRow from "./DashboardsTagRow.svelte";

  /** All dashboards (unfiltered). Used to derive the available tags and counts. */
  export let resources: V1Resource[] = [];
  /** The currently selected tag, or null when showing all. Bindable. */
  export let selectedTag: string | null = null;

  let searchText = "";

  // Derive the tag rows (name + count) from the dashboards' meta tags. The tag
  // index dedupes and counts; we sort alphabetically for a stable listing.
  $: tags = buildTagIndex(
    resources.map((r) => ({
      name: r.meta?.name?.name,
      tags: r.meta?.tags,
    })),
  ).tags.sort((a, b) => a.name.localeCompare(b.name));

  // Drop the selection if the active tag disappears (e.g. the YAML was edited).
  $: if (selectedTag && !tags.some((t) => t.name === selectedTag)) {
    selectedTag = null;
  }

  $: filteredTags = searchText.trim()
    ? tags.filter((t) =>
        t.name.toLowerCase().includes(searchText.trim().toLowerCase()),
      )
    : tags;

  function toggleTag(name: string) {
    selectedTag = selectedTag === name ? null : name;
  }
</script>

<div class="sidebar">
  <div class="input-wrapper">
    <Search
      bind:value={searchText}
      placeholder="Search tags"
      label="Search tags"
      autofocus={false}
      showBorderOnFocus={false}
    />
  </div>

  <div class="tags-scroll">
    <h3 class="column-header">Tags</h3>

    {#if filteredTags.length === 0}
      <p class="text-fg-secondary my-1 px-2 text-xs">No matching tags</p>
    {:else}
      {#each filteredTags as tag (tag.name)}
        <DashboardsTagRow
          name={tag.name}
          count={tag.totalCount}
          selected={selectedTag === tag.name}
          onSelect={() => toggleTag(tag.name)}
        />
      {/each}
    {/if}
  </div>
</div>

<style lang="postcss">
  .sidebar {
    @apply flex flex-col overflow-hidden;
    @apply h-full w-full border-r bg-surface-background;
    @apply select-none;
  }

  .input-wrapper {
    @apply flex w-full h-fit items-center border-b gap-x-2 p-2;
  }

  .tags-scroll {
    @apply flex flex-col h-full w-full py-2 px-2 gap-y-0.5;
    @apply overflow-y-auto overflow-x-hidden;
  }

  .column-header {
    @apply uppercase font-semibold text-[10px];
    @apply px-1.5 pt-1 pb-1 text-fg-secondary;
  }
</style>
