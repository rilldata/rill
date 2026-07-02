<script lang="ts">
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import {
    getResourceTags,
    useDashboards,
  } from "@rilldata/web-admin/features/dashboards/listing/selectors.ts";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { UrlParamsArrayState } from "@rilldata/web-common/lib/url-params-state.svelte.ts";

  let {
    align = "start",
  }: {
    align?: "start" | "end";
  } = $props();

  let open = $state(false);

  const selectedTagsState = UrlParamsArrayState.createStringArrayParam("tags");

  const runtimeClient = useRuntimeClient();
  let dashboards = useDashboards(runtimeClient);
  let availableTags = $derived(
    Array.from(
      new Set($dashboards?.data?.flatMap(getResourceTags) ?? []),
    ).sort(),
  );

  let tagsLabel = $derived(
    selectedTagsState.value.length === 0
      ? "All tags"
      : selectedTagsState.value.length === 1
        ? selectedTagsState.value[0]
        : `${selectedTagsState.value[0]}, +${selectedTagsState.value.length - 1} other${selectedTagsState.value.length > 2 ? "s" : ""}`,
  );
</script>

{#if availableTags.length > 0}
  <DropdownMenu.Root bind:open>
    <DropdownMenu.Trigger
      class="min-w-fit min-h-7 flex flex-row gap-1 items-center rounded-sm border bg-input {open
        ? 'bg-gray-200'
        : 'hover:bg-surface-hover'} px-2 py-1"
    >
      <span class="text-fg-secondary font-medium text-sm">{tagsLabel}</span>
      {#if open}
        <CaretUpIcon size="12px" />
      {:else}
        <CaretDownIcon size="12px" />
      {/if}
    </DropdownMenu.Trigger>
    <DropdownMenu.Content {align} class="w-48 max-h-72 overflow-y-auto">
      {#each availableTags as tag (tag)}
        <DropdownMenu.CheckboxItem
          checked={selectedTagsState.value.includes(tag)}
          onCheckedChange={() => selectedTagsState.toggle(tag)}
        >
          {tag}
        </DropdownMenu.CheckboxItem>
      {/each}
    </DropdownMenu.Content>
  </DropdownMenu.Root>
{/if}
