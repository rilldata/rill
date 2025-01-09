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

  $: selected = options.get(current.toLowerCase());

  let groupedData = new Map();

  for (let [key] of options) {
    const [group, sub] = key.split(":");
    if (sub) {
      if (!groupedData.has(group)) {
        groupedData.set(group, []);
      }
      groupedData.get(group).push(sub);
    } else {
      if (!groupedData.has(group)) {
        groupedData.set(group, null); // options without a namespace
      }
    }
  }

  $: console.log(groupedData);
</script>

<li class="flex items-center gap-x-2 px-2">
  <div class="flex flex-row gap-x-1 items-center">
    {#if options.size > 1}
      <DropdownMenu.Root>
        <DropdownMenu.Trigger asChild let:builder>
          <button use:builder.action {...builder} class="trigger">
            <CaretDownIcon size="14px" />
          </button>
        </DropdownMenu.Trigger>
        <DropdownMenu.Content align="start" class="min-w-44 max-h-96">
          {#each Array.from(groupedData.entries()) as [group, subItems]}
            {#if subItems}
              <DropdownMenu.Sub>
                <DropdownMenu.SubTrigger>Group Name</DropdownMenu.SubTrigger>
                <DropdownMenu.SubContent
                  align="start"
                  class="min-w-44 max-h-96"
                >
                  {#each subItems as subItem}
                    <DropdownMenu.CheckboxItem
                      >{subItem}</DropdownMenu.CheckboxItem
                    >
                  {/each}
                </DropdownMenu.SubContent>
              </DropdownMenu.Sub>
            {:else}
              <DropdownMenu.CheckboxItem>{group}</DropdownMenu.CheckboxItem>
            {/if}
          {/each}
        </DropdownMenu.Content>
      </DropdownMenu.Root>
    {/if}
    {#if selected}
      <a
        href={isCurrentPage ? "#top" : undefined}
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
