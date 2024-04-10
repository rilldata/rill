<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
  import { page } from "$app/stores";
  import type { Level } from "./Breadcrumbs.svelte";

  const regex = /\[.*?\]/g;

  export let options: Level;
  export let current: string;
  export let isCurrentPage = false;
  export let depth: number;

  $: selected = options.get(current);
  $: nextChild = $page.route.id?.split(regex)[depth + 1];
</script>

<li class="flex items-center gap-x-2 px-2">
  <div class="flex flex-row gap-x-1 items-center">
    {#if selected}
      <a
        href={isCurrentPage ? "#top" : selected?.href + nextChild}
        class="text-gray-500 hover:text-gray-600"
        class:current={isCurrentPage}
      >
        {selected?.label}
      </a>
    {/if}
    {#if options.size > 1}
      <DropdownMenu.Root>
        <DropdownMenu.Trigger asChild let:builder>
          <button use:builder.action {...builder} class="trigger">
            <CaretDownIcon size="14px" />
          </button>
        </DropdownMenu.Trigger>
        <DropdownMenu.Content align="start" class="max-h-96 overflow-auto">
          {#each options as [id, { label, href }] (id)}
            {@const selected = id === current}
            <DropdownMenu.Item {href}>
              <div class="item" class:pl-4={!selected}>
                <Check className={!selected ? "hidden" : ""} />

                <svelte:element this={selected ? "b" : "span"}>
                  {label}
                </svelte:element>
              </div>
            </DropdownMenu.Item>
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

  .item {
    @apply text-gray-800 flex gap-x-2 items-center;
  }

  .trigger {
    @apply flex flex-col justify-center items-center;
    @apply transition-transform  text-gray-500;
  }

  .trigger:hover {
    @apply translate-y-[2px];
  }
</style>
