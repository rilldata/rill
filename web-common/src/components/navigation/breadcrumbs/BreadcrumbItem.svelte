<script lang="ts">
  import { page } from "$app/stores";
  import { Chip } from "@rilldata/web-common/components/chip";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { getNonVariableSubRoute } from "@rilldata/web-common/components/navigation/breadcrumbs/utils";
  import {
    resourceColorMapping,
    resourceIconMapping,
  } from "@rilldata/web-common/features/entity-management/resource-icon-mapping";
  import type { PathOption, PathOptions } from "./types";

  export let options: PathOptions;
  export let current: string;
  export let isCurrentPage = false;
  export let depth: number = 0;
  export let currentPath: (string | undefined)[] = [];
  export let onSelect: undefined | ((id: string) => void) = undefined;
  export let isEmbedded: boolean = false;

  const DEFAULT_GROUP_ORDER = 1000;

  $: normalizedCurrent = current.toLowerCase();
  $: selected = options.get(normalizedCurrent);
  $: selectedIconComponent =
    selected?.resourceKind && resourceIconMapping[selected.resourceKind];
  $: selectedIconColor =
    selected?.resourceKind && resourceColorMapping[selected.resourceKind];

  type OptionEntry = [string, PathOption];

  type GroupedOptions = {
    id: string;
    label: string;
    order: number;
    options: OptionEntry[];
  };

  function partitionOptions(
    options: PathOptions,
    selectedGroupId?: string,
  ): { grouped: GroupedOptions[]; ungrouped: OptionEntry[] } {
    const groups = new Map<string, GroupedOptions>();
    const ungrouped: OptionEntry[] = [];

    options.forEach((option, id) => {
      const groupId = option.groupId ?? option.groupLabel?.toLowerCase();
      if (groupId) {
        const label = option.groupLabel ?? option.groupId ?? groupId;
        const order = option.groupOrder ?? DEFAULT_GROUP_ORDER;
        const existing = groups.get(groupId) ?? {
          id: groupId,
          label,
          order,
          options: [],
        };
        existing.label = label;
        existing.order = Math.min(existing.order, order);
        existing.options.push([id, option]);
        groups.set(groupId, existing);
      } else {
        ungrouped.push([id, option]);
      }
    });

    const grouped = Array.from(groups.values()).sort((a, b) => {
      if (selectedGroupId) {
        if (a.id === selectedGroupId && b.id !== selectedGroupId) return -1;
        if (b.id === selectedGroupId && a.id !== selectedGroupId) return 1;
      }
      if (a.order !== b.order) return a.order - b.order;
      return a.label.localeCompare(b.label);
    });

    grouped.forEach((group) => {
      group.options.sort((a, b) => a[1].label.localeCompare(b[1].label));
    });

    ungrouped.sort((a, b) => a[1].label.localeCompare(b[1].label));

    return { grouped, ungrouped };
  }

  $: selectedGroupId = selected?.groupId ?? selected?.groupLabel?.toLowerCase();
  $: partitionedOptions = partitionOptions(options, selectedGroupId);
  $: hasGroups = partitionedOptions.grouped.length > 0;

  function linkMaker(
    current: (string | undefined)[],
    depth: number,
    id: string,
    option: PathOption,
    route: string,
  ) {
    if (onSelect) return undefined;
    if (option?.href) return option.href;

    const newPath = current
      .slice(0, option?.depth ?? depth)
      .filter((p): p is string => !!p);

    if (option?.section) newPath.push(option.section);

    newPath.push(id);
    const path = `/${newPath.join("/")}`;

    // add the sub route if it has no variables
    return path + getNonVariableSubRoute(path, route);
  }
</script>

<li class="flex items-center gap-x-2 px-2">
  <div
    class="flex flex-row gap-x-1 items-center"
    aria-label="Breadcrumb navigation, level {depth}"
  >
    {#if selected}
      <a
        on:click={() => {
          if (isCurrentPage && !isEmbedded) window.location.reload();
        }}
        href={isCurrentPage
          ? "#top"
          : linkMaker(currentPath, depth, current, selected, "")}
        class="text-gray-500 hover:text-gray-600 flex flex-row items-center gap-x-1.5"
        class:current={isCurrentPage}
      >
        {#if selectedIconComponent}
          <svelte:component
            this={selectedIconComponent}
            size="14"
            color={selectedIconColor}
            class="shrink-0 text-gray-500"
          />
        {/if}
        <span>{selected?.label}</span>
      </a>
      {#if selected?.pill}
        <Chip type="dimension" label={selected.pill} readOnly compact>
          <svelte:fragment slot="body">{selected.pill}</svelte:fragment>
        </Chip>
      {/if}
    {/if}
    {#if options.size > 1}
      <DropdownMenu.Root>
        <DropdownMenu.Trigger asChild let:builder>
          <button
            use:builder.action
            {...builder}
            class="trigger"
            aria-label="Breadcrumb dropdown"
          >
            <CaretDownIcon size="14px" />
          </button>
        </DropdownMenu.Trigger>
        <DropdownMenu.Content
          align="start"
          class="min-w-44 max-h-96 overflow-y-auto"
        >
          {#if hasGroups}
            {#each partitionedOptions.grouped as group, groupIndex}
              <DropdownMenu.Label
                class="px-2 pt-1 pb-0 text-[10px] font-medium uppercase text-gray-400 tracking-wide"
              >
                {group.label}
              </DropdownMenu.Label>
              {#each group.options as [id, option] (id)}
                {@const checked = id === normalizedCurrent}
                {@const IconComponent = option.resourceKind
                  ? resourceIconMapping[option.resourceKind]
                  : null}
                {@const iconColor = option.resourceKind
                  ? resourceColorMapping[option.resourceKind]
                  : undefined}
                <DropdownMenu.CheckboxItem
                  class="cursor-pointer"
                  {checked}
                  checkSize={"h-3 w-3"}
                  href={linkMaker(
                    currentPath,
                    depth,
                    id,
                    option,
                    $page.route.id ?? "",
                  )}
                  preloadData={option.preloadData}
                  on:click={() => {
                    if (onSelect) {
                      onSelect(id);
                    }
                  }}
                >
                  <span
                    class="text-xs text-gray-800 flex-grow flex items-center gap-x-2"
                  >
                    {#if IconComponent}
                      <IconComponent
                        size="14"
                        color={iconColor}
                        class="shrink-0 text-gray-500"
                      />
                    {/if}
                    <span class="truncate">{option.label}</span>
                  </span>
                </DropdownMenu.CheckboxItem>
              {/each}
              {#if groupIndex < partitionedOptions.grouped.length - 1}
                <DropdownMenu.Separator />
              {/if}
            {/each}
            {#if partitionedOptions.ungrouped.length}
              <DropdownMenu.Separator />
            {/if}
            {#each partitionedOptions.ungrouped as [id, option] (id)}
              {@const checked = id === normalizedCurrent}
              {@const IconComponent = option.resourceKind
                ? resourceIconMapping[option.resourceKind]
                : null}
              {@const iconColor = option.resourceKind
                ? resourceColorMapping[option.resourceKind]
                : undefined}
              <DropdownMenu.CheckboxItem
                class="cursor-pointer"
                {checked}
                checkSize={"h-3 w-3"}
                href={linkMaker(
                  currentPath,
                  depth,
                  id,
                  option,
                  $page.route.id ?? "",
                )}
                preloadData={option.preloadData}
                on:click={() => {
                  if (onSelect) {
                    onSelect(id);
                  }
                }}
              >
                <span
                  class="text-xs text-gray-800 flex-grow flex items-center gap-x-2"
                >
                  {#if IconComponent}
                    <IconComponent
                      size="14"
                      color={iconColor}
                      class="shrink-0 text-gray-500"
                    />
                  {/if}
                  <span class="truncate">{option.label}</span>
                </span>
              </DropdownMenu.CheckboxItem>
            {/each}
          {:else}
            {#each options as [id, option] (id)}
              {@const checked = id === normalizedCurrent}
              {@const IconComponent = option.resourceKind
                ? resourceIconMapping[option.resourceKind]
                : null}
              {@const iconColor = option.resourceKind
                ? resourceColorMapping[option.resourceKind]
                : undefined}
              <DropdownMenu.CheckboxItem
                class="cursor-pointer"
                {checked}
                checkSize={"h-3 w-3"}
                href={linkMaker(
                  currentPath,
                  depth,
                  id,
                  option,
                  $page.route.id ?? "",
                )}
                preloadData={option.preloadData}
                on:click={() => {
                  if (onSelect) {
                    onSelect(id);
                  }
                }}
              >
                <span
                  class="text-xs text-gray-800 flex-grow flex items-center gap-x-2"
                >
                  {#if IconComponent}
                    <IconComponent
                      size="14"
                      color={iconColor}
                      class="shrink-0 text-gray-500"
                    />
                  {/if}
                  <span class="truncate">{option.label}</span>
                </span>
              </DropdownMenu.CheckboxItem>
            {/each}
          {/if}
        </DropdownMenu.Content>
      </DropdownMenu.Root>
    {/if}
  </div>
</li>

<style lang="postcss">
  .current {
    @apply text-gray-800 font-medium;
  }

  .trigger {
    @apply flex flex-col justify-center items-center;
    @apply transition-transform  text-gray-500;
    @apply px-0.5 py-1 rounded;
  }

  .trigger:hover,
  .trigger[data-state="open"] {
    @apply bg-gray-100;
  }
</style>
