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
    isEmbedded = false,
  }: {
    pathOptions: PathOptions;
    current: string;
    isCurrentPage?: boolean;
    depth?: number;
    currentPath?: (string | undefined)[];
    onSelect?: (id: string) => void;
    isEmbedded?: boolean;
  } = $props();

  let options = $derived(pathOptions.options);
  let carryOverSearchParams = $derived(pathOptions.carryOverSearchParams);
  let content = $derived(pathOptions.content);
  let selected = $derived(options.get(current.toLowerCase()));

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
    return path + getCarryOverSubRoute(route, path);
  }
</script>

<li class="flex items-center gap-x-2 px-2">
  <div
    class="flex flex-row gap-x-1 items-center"
    aria-label="Breadcrumb navigation, level {depth}"
  >
    {#if selected}
      <a
        onclick={() => {
          if (isCurrentPage && !isEmbedded) window.location.reload();
        }}
        href={isCurrentPage
          ? "#top"
          : linkMaker(currentPath, depth, current, selected, "")}
        class={[
          "text-fg-muted hover:text-fg-secondary flex flex-row items-center gap-x-2",
          { current: isCurrentPage },
        ]}
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
        <DropdownMenu.Trigger>
          {#snippet child({ props })}
            <button {...props} class="trigger" aria-label="Breadcrumb dropdown">
              <CaretDownIcon size="14px" />
            </button>
          {/snippet}
        </DropdownMenu.Trigger>

        {#if content}
          {@render content({
            options,
            current,
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
                {current}
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
</style>
