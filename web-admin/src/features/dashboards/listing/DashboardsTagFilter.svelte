<script lang="ts">
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { useDashboards } from "@rilldata/web-admin/features/dashboards/listing/selectors.ts";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import {
    getAllTagsForResources,
    getTagFilterLabel,
  } from "@rilldata/web-common/features/resources/resource-tag-utils.ts";
  import type { ArrayRuneStore } from "web-common/src/lib/store-utils/types.svelte.ts";
  import {
    getDashboardTagFavouritesStore,
    sortByFavourites,
  } from "./dashboard-favourites.ts";
  import { page } from "$app/state";

  let {
    align = "start",
    size = "sm",
    selectedTagsStore,
  }: {
    align?: "start" | "end";
    size?: "xs" | "sm";
    selectedTagsStore: ArrayRuneStore<string>;
  } = $props();

  let open = $state(false);

  const runtimeClient = useRuntimeClient();
  let dashboards = useDashboards(runtimeClient);
  let availableTags = $derived(getAllTagsForResources($dashboards?.data ?? []));

  let tagsLabel = $derived(getTagFilterLabel(selectedTagsStore.value));

  let { organization, project } = $derived(page.params);
  let tagsFavourites = $derived(
    getDashboardTagFavouritesStore(organization, project),
  );

  let sortedTags = $derived(
    sortByFavourites(availableTags, tagsFavourites.value, (t) => t.name),
  );
</script>

{#if availableTags.length > 0}
  <DropdownMenu.Root bind:open>
    <DropdownMenu.Trigger
      class="min-w-fit min-h-7 flex flex-row gap-1 items-center rounded-sm border bg-input {open
        ? 'bg-gray-200'
        : 'hover:bg-surface-hover'} px-2 py-1"
    >
      <span class="text-fg-secondary font-medium text-{size}">{tagsLabel}</span>
      {#if open}
        <CaretUpIcon size="12px" />
      {:else}
        <CaretDownIcon size="12px" />
      {/if}
    </DropdownMenu.Trigger>
    <DropdownMenu.Content {align} class="w-48 max-h-72 overflow-y-auto">
      {#each sortedTags as tag (tag.name)}
        <DropdownMenu.CheckboxItem
          checked={selectedTagsStore.value.includes(tag.name)}
          onCheckedChange={() => selectedTagsStore.toggle(tag.name)}
        >
          {tag.name}
        </DropdownMenu.CheckboxItem>
      {/each}
    </DropdownMenu.Content>
  </DropdownMenu.Root>
{/if}
