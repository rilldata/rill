<script lang="ts">
  import { page } from "$app/stores";
  import { Chip } from "@rilldata/web-common/components/chip";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import type { PathOption, PathOptions } from "./types";
  import { getNonVariableSubRoute } from "@rilldata/web-common/components/navigation/breadcrumbs/utils.ts";
  import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params.ts";
  import { resourceIconMapping } from "@rilldata/web-common/features/entity-management/resource-icon-mapping";

  export let pathOptions: PathOptions;
  export let current: string;
  export let isCurrentPage = false;
  export let depth: number = 0;
  export let currentPath: (string | undefined)[] = [];
  export let onSelect: undefined | ((id: string) => void) = undefined;
  export let isEmbedded: boolean = false;

  $: ({ options, groups, carryOverSearchParams } = pathOptions);
  $: hasGroups = groups && groups.length > 0;
  $: selected = options.get(current.toLowerCase());

  function linkMaker(
    current: (string | undefined)[],
    depth: number,
    id: string,
    option: PathOption,
    route: string,
  ) {
    const path = makePath(current, depth, id, option, route);

    if (!path || !carryOverSearchParams || $page.url.search === "") return path;

    const url = new URL($page.url);
    url.pathname = path;
    url.searchParams.set(ExploreStateURLParams.IgnoreErrors, "true");
    return url.pathname + url.search;
  }

  function makePath(
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
        class="text-fg-muted hover:text-fg-secondary flex flex-row items-center gap-x-2"
        class:current={isCurrentPage}
      >
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
            {#each groups as group, groupIndex (group.name)}
              <DropdownMenu.Group class="px-1">
                {#if groups.length > 1}
                  <DropdownMenu.Label>
                    {group.label}
                  </DropdownMenu.Label>
                {/if}
                {#each group.items as { id, option } (id)}
                  {@const isSelected = id === current.toLowerCase()}
                  {@const icon = option.resourceKind
                    ? resourceIconMapping[option.resourceKind]
                    : undefined}
                  <DropdownMenu.CheckboxItem
                    class="cursor-pointer"
                    checked={isSelected}
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
                      class="text-xs text-fg-secondary flex-grow flex items-center gap-x-1.5"
                    >
                      {#if icon}
                        <svelte:component this={icon} size="12px" />
                      {/if}
                      {option.label}
                    </span>
                  </DropdownMenu.CheckboxItem>
                {/each}
                {#if groupIndex !== groups.length - 1}
                  <DropdownMenu.Separator />
                {/if}
              </DropdownMenu.Group>
            {/each}
          {:else}
            {#each options as [id, option] (id)}
              {@const isSelected = id === current.toLowerCase()}
              {@const icon = option.resourceKind
                ? resourceIconMapping[option.resourceKind]
                : undefined}
              <DropdownMenu.CheckboxItem
                class="cursor-pointer"
                checked={isSelected}
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
                  class="text-xs text-fg-secondary flex-grow flex items-center gap-x-1.5"
                >
                  {#if icon}
                    <svelte:component this={icon} size="12px" />
                  {/if}
                  {option.label}
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
    @apply text-fg-muted font-medium;
  }

  .trigger {
    @apply flex flex-col justify-center items-center;
    @apply transition-transform text-fg-muted;
    @apply px-0.5 py-1 rounded;
  }

  .trigger:hover,
  .trigger[data-state="open"] {
    @apply bg-gray-100;
  }
</style>
