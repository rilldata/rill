<script lang="ts">
  import { page } from "$app/stores";
  import { Chip } from "@rilldata/web-common/components/chip";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { getNonVariableSubRoute } from "@rilldata/web-common/components/navigation/breadcrumbs/utils";
  import type { PathOption, PathOptions } from "./types";

  export let options: PathOptions;
  export let current: string;
  export let isCurrentPage = false;
  export let depth: number = 0;
  export let currentPath: (string | undefined)[] = [];
  export let onSelect: undefined | ((id: string) => void) = undefined;
  export let isEmbedded: boolean = false;

  $: selected = options.get(current.toLowerCase());

  let groupedData = new Map();

  // Group data by colon separator
  for (let [key, value] of options) {
    const [group, sub] = key.split(":");
    if (sub) {
      if (!groupedData.has(group)) {
        groupedData.set(group, []);
      }
      groupedData.get(group).push({ key, label: value.label });
    } else {
      groupedData.set(group, { key, label: value.label }); // Standalone items
    }
  }

  const handleSelect = (key: string) => {
    if (onSelect) {
      onSelect(key);
    }
  };

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

{#if !isEmbedded}
  <li class="flex items-center gap-x-2 px-2">
    <div class="flex flex-row gap-x-1 items-center">
      {#if selected}
        <a
          on:click={() => {
            if (isCurrentPage && !isEmbedded) window.location.reload();
          }}
          href={isCurrentPage
            ? "#top"
            : linkMaker(currentPath, depth, current, selected, "")}
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
            <button use:builder.action {...builder} class="trigger">
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
          </DropdownMenu.Content>
        </DropdownMenu.Root>
      {/if}
    </div>
  </li>
{:else}
  <li class="flex items-center gap-x-2 px-2">
    <div class="flex flex-row gap-x-1 items-center">
      {#if selected}
        <a
          href={isCurrentPage ? "#top" : undefined}
          class="text-gray-500 hover:text-gray-600 flex flex-row items-center gap-x-2"
          class:current={isCurrentPage}
        >
          <span>{selected?.label}</span>
        </a>
      {/if}
      {#if options.size > 1}
        <DropdownMenu.Root>
          <DropdownMenu.Trigger asChild let:builder>
            <button use:builder.action {...builder} class="trigger">
              <CaretDownIcon size="14px" />
            </button>
          </DropdownMenu.Trigger>
          <DropdownMenu.Content align="start" class="min-w-44 max-h-96">
            {#each Array.from(groupedData.entries()) as [group, subItems]}
              {#if Array.isArray(subItems)}
                <!-- Grouped submenu -->
                <DropdownMenu.Sub>
                  <DropdownMenu.SubTrigger>{group}</DropdownMenu.SubTrigger>
                  <DropdownMenu.SubContent
                    align="start"
                    class="min-w-44 max-h-96"
                  >
                    {#each subItems as subItem}
                      <DropdownMenu.Item
                        class="cursor-pointer"
                        on:click={() => handleSelect(subItem.key)}
                      >
                        <span class="text-xs text-gray-800 flex-grow">
                          {subItem.label}
                        </span>
                      </DropdownMenu.Item>
                    {/each}
                  </DropdownMenu.SubContent>
                </DropdownMenu.Sub>
              {:else}
                <!-- Standalone item -->
                <DropdownMenu.Item
                  class="cursor-pointer"
                  on:click={() => handleSelect(subItems.key)}
                >
                  <span class="text-xs text-gray-800 flex-grow">
                    {subItems.label}
                  </span>
                </DropdownMenu.Item>
              {/if}
            {/each}
          </DropdownMenu.Content>
        </DropdownMenu.Root>
      {/if}
    </div>
  </li>
{/if}

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
