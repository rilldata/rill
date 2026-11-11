<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import type { PathOptions } from "./types";

  export let pathOptions: PathOptions;
  export let current: string;
  export let isCurrentPage = false;

  $: ({ options } = pathOptions);

  $: selected = options.get(current.toLowerCase());
  $: hasDropdown = options.size > 1;
  // The current page's label would only link back to itself,
  // so the whole crumb opens the dropdown instead of just the chevron.
  $: isWholeCrumbTrigger = isCurrentPage && hasDropdown && !!selected;

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
    {#if selected && !isWholeCrumbTrigger}
      <span
        class="text-fg-secondary"
        class:current={isCurrentPage}
        aria-current={isCurrentPage ? "page" : undefined}
      >
        {selected.label}
      </span>
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
                <span class="current">{selected?.label}</span>
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
                      data-sveltekit-preload-data="false"
                    >
                      <span class="text-xs text-fg-primary flex-grow">
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
                data-sveltekit-preload-data="false"
              >
                <span class="text-xs text-fg-primary flex-grow">
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
    @apply text-fg-primary font-medium;
  }

  .trigger {
    @apply flex flex-col justify-center items-center;
    @apply transition-transform text-fg-secondary;
    @apply px-0.5 py-1 rounded;
  }

  .trigger:hover {
    @apply bg-gray-100;
  }

  .trigger:global([data-state="open"]) {
    @apply bg-gray-100;
  }

  /* Keeps the label and chevron where they sit when they are separate controls. */
  .whole-crumb {
    @apply flex-row gap-x-1.5 pl-1 pr-0.5 -ml-1;
  }
</style>
