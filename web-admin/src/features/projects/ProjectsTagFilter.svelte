<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";
  import { normalizeProjectTag } from "@rilldata/web-admin/features/projects/project-tags";
  import { getTagFilterLabel } from "@rilldata/web-common/features/resources/resource-tag-utils.ts";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { cn } from "@rilldata/web-common/lib/shadcn";
  import type { ArrayRuneStore } from "@rilldata/web-common/lib/store-utils/types.svelte.ts";

  let {
    tags,
    selectedTagsStore,
  }: {
    tags: string[];
    selectedTagsStore: ArrayRuneStore<string>;
  } = $props();

  let open = $state(false);

  // Tags in the URL match regardless of case, like the admin service's tags filter, so `?tags=Prod` selects `prod`.
  let selected = $derived([
    ...new Set(selectedTagsStore.value.map(normalizeProjectTag)),
  ]);
  // Selected tags stay in the list even if no project has them anymore (e.g. from an old link), so they can be cleared.
  let options = $derived(
    [...new Set([...tags, ...selected])].sort((a, b) => a.localeCompare(b)),
  );
  let label = $derived(getTagFilterLabel(selected));

  function toggle(tag: string) {
    selectedTagsStore.setter(
      selected.includes(tag)
        ? selected.filter((t) => t !== tag)
        : [...selected, tag],
    );
  }
</script>

<DropdownMenu.Root bind:open>
  <DropdownMenu.Trigger
    class={cn(
      "flex flex-row items-center gap-x-2 h-7 px-3 rounded-[2px] border text-xs font-medium",
      open
        ? "bg-surface-hover text-fg-accent"
        : "text-fg-secondary hover:bg-surface-hover",
    )}
    aria-label={m.org_projects_tag_filter_label()}
  >
    <span class="max-w-48 truncate">{label}</span>
    {#if open}
      <CaretUpIcon size="12px" />
    {:else}
      <CaretDownIcon size="12px" />
    {/if}
  </DropdownMenu.Trigger>
  <DropdownMenu.Content align="end" class="w-56 max-h-72 overflow-y-auto">
    {#each options as tag (tag)}
      <DropdownMenu.CheckboxItem
        closeOnSelect={false}
        checked={selected.includes(tag)}
        onCheckedChange={() => toggle(tag)}
      >
        <span class="truncate">{tag}</span>
      </DropdownMenu.CheckboxItem>
    {/each}
    {#if selected.length > 0}
      <DropdownMenu.Separator />
      <DropdownMenu.Item onclick={() => selectedTagsStore.setter([])}>
        {m.org_projects_tag_filter_clear()}
      </DropdownMenu.Item>
    {/if}
  </DropdownMenu.Content>
</DropdownMenu.Root>
