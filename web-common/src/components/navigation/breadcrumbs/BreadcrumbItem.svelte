<script lang="ts">
  import { page } from "$app/stores";
  import { Chip } from "@rilldata/web-common/components/chip";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import type { PathOption, PathOptions } from "./types";
  import { getCarryOverSubRoute } from "@rilldata/web-common/components/navigation/breadcrumbs/utils.ts";
  import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params.ts";
  import BreadcrumbDropdownItem from "@rilldata/web-common/components/navigation/breadcrumbs/BreadcrumbDropdownItem.svelte";

  let {
    pathOptions,
    current,
    isCurrentPage = false,
    depth = 0,
    currentPath = [],
    onSelect = undefined,
  }: {
    pathOptions: PathOptions;
    current: string;
    isCurrentPage?: boolean;
    depth?: number;
    currentPath?: (string | undefined)[];
    onSelect?: (id: string) => void;
  } = $props();

  let options = $derived(pathOptions.options);
  let carryOverSearchParams = $derived(pathOptions.carryOverSearchParams);
  let content = $derived(pathOptions.content);
  let currentId = $derived(pathOptions.currentId ?? current.toLowerCase());
  let selected = $derived(options.get(currentId));
  let hasDropdown = $derived(options.size > 1);
  // The current page's label would only link back to itself,
  // so the whole crumb opens the dropdown instead of just the chevron.
  let isWholeCrumbTrigger = $derived(
    isCurrentPage && hasDropdown && !!selected,
  );

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

    newPath.push(option?.param ?? id);
    const path = `/${newPath.join("/")}`;
    return path + getCarryOverSubRoute(route, path);
  }
</script>

{#snippet pill(option: PathOption)}
  {#if option.pill}
    <Chip type="dimension" label={option.pill} readOnly compact>
      <svelte:fragment slot="body">{option.pill}</svelte:fragment>
    </Chip>
  {/if}
{/snippet}

<li class="flex items-center gap-x-2 px-2">
  <div
    class="flex flex-row gap-x-1 items-center"
    aria-label="Breadcrumb navigation, level {depth}"
  >
    {#if selected && !isWholeCrumbTrigger}
      {#if isCurrentPage}
        <span class="current" aria-current="page">{selected.label}</span>
      {:else}
        <a
          href={linkMaker(currentPath, depth, currentId, selected, "")}
          class="text-fg-muted hover:text-fg-secondary flex flex-row items-center gap-x-2"
        >
          <span>{selected.label}</span>
        </a>
      {/if}
      {@render pill(selected)}
    {/if}
    {#if hasDropdown}
      <DropdownMenu.Root>
        <DropdownMenu.Trigger>
          {#snippet child({ props })}
            {#if isWholeCrumbTrigger}
              <button
                {...props}
                class="trigger whole-crumb"
                aria-current="page"
              >
                {#if selected}
                  <span class="current">{selected.label}</span>
                  {@render pill(selected)}
                {/if}
                <CaretDownIcon size="14px" />
              </button>
            {:else}
              <button
                {...props}
                class="trigger"
                aria-label="Breadcrumb dropdown"
              >
                <CaretDownIcon size="14px" />
              </button>
            {/if}
          {/snippet}
        </DropdownMenu.Trigger>

        {#if content}
          {@render content({
            options,
            current: currentId,
            currentPath,
            depth,
            onSelect,
            linkMaker,
          })}
        {:else}
          <DropdownMenu.Content
            align="start"
            class="min-w-44 max-h-96 overflow-y-auto"
          >
            {#each options as [id, option] (id)}
              <BreadcrumbDropdownItem
                {id}
                {option}
                current={currentId}
                {currentPath}
                {depth}
                {onSelect}
                {linkMaker}
              />
            {/each}
          </DropdownMenu.Content>
        {/if}
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
  .trigger:global([data-state="open"]) {
    @apply bg-gray-100;
  }

  /* Keeps the label and chevron where they sit when they are separate controls. */
  .whole-crumb {
    @apply flex-row gap-x-1.5 pl-1 pr-0.5 -ml-1;
  }

  .whole-crumb:hover .current {
    @apply text-fg-secondary;
  }
</style>
