<script lang="ts">
  import { page } from "$app/stores";
  import { Chip } from "@rilldata/web-common/components/chip";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { makePath } from "@rilldata/web-common/components/navigation/breadcrumbs/utils";
  import type { PathOption, PathOptions } from "./types";
  import { get } from "svelte/store";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";

  export let options: PathOptions["options"];
  export let current: string;
  export let isCurrentPage = false;
  export let depth: number = 0;
  export let currentPath: (string | undefined)[] = [];
  export let onSelect: undefined | ((id: string) => void) = undefined;
  export let isEmbedded: boolean = false;

  $: selected = options.get(current.toLowerCase());

  let carryOverSearchParams = false;

  function linkMaker(
    current: (string | undefined)[],
    depth: number,
    id: string,
    option: PathOption,
    route: string,
    carryOverSearchParams: boolean, // needed for reactivity
  ) {
    const path = makePath(current, depth, id, option, route, !!onSelect);
    if (!path) return undefined;

    if (!carryOverSearchParams) return path;

    const url = new URL(window.location.href);
    if (url.search === "") return path;

    url.pathname = path;
    url.searchParams.set("ignore_errors", "true");
    return url.pathname + url.search;
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
          : linkMaker(
              currentPath,
              depth,
              current,
              selected,
              "",
              carryOverSearchParams,
            )}
        class="text-gray-500 hover:text-gray-600 flex flex-row items-center gap-x-2"
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
          {#each options as [id, option] (id)}
            {@const selected = id === current.toLowerCase()}
            <DropdownMenu.CheckboxItem
              class="cursor-pointer"
              checked={selected}
              checkSize={"h-3 w-3"}
              href={linkMaker(
                currentPath,
                depth,
                id,
                option,
                $page.route.id ?? "",
                carryOverSearchParams,
              )}
              preloadData={option.preloadData}
              on:click={() => {
                if (onSelect) {
                  onSelect(id);
                }
              }}
            >
              <span class="text-xs text-gray-800 flex-grow">
                {option.label}
              </span>
            </DropdownMenu.CheckboxItem>
          {/each}

          <div class="flex flex-row items-center gap-x-2 pt-1 border-t">
            <Switch small bind:checked={carryOverSearchParams} />
            Carry over dashboard state
          </div>
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
