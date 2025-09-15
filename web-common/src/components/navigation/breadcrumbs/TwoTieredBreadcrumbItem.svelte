<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import type { PathOptions } from "./types";

  export let options: PathOptions;
  export let current: string;
  export let isCurrentPage = false;

  $: selected = options.get(current.toLowerCase());

  const groupedData = new Map();

  // Group data by colon separator
  for (let [key, option] of options) {
    const [group, sub] = key.split(":");
    if (sub) {
      if (!groupedData.has(group)) {
        groupedData.set(group, []);
      }
      groupedData.get(group).push(option);
    } else {
      groupedData.set(group, option); // Standalone items
    }
  }
</script>

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
        <DropdownMenu.Content align="start" class="min-w-44">
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
                      href={subItem.href}
                      preloadData={false}
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
                href={subItems.href}
                preloadData={false}
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

<style lang="postcss">
  .current {
    @apply text-gray-800 font-medium;
  }

  .trigger {
    @apply flex flex-col justify-center items-center;
    @apply transition-transform text-gray-500;
    @apply px-0.5 py-1 rounded;
  }

  .trigger:hover,
  .trigger[data-state="open"] {
    @apply bg-gray-100;
  }
</style>
