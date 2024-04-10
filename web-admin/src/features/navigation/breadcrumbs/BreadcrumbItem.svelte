<script lang="ts" context="module">
  export interface BreadcrumbMenuItem {
    key: string;
    main: string;
    kind?: ResourceKind;
  }
</script>

<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
  import Spacer from "@rilldata/web-common/components/icons/Spacer.svelte";
  import type { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";

  export let label: string;
  export let href: string;
  export let menuItems: BreadcrumbMenuItem[] = [];
  export let menuKey: string;
  export let makeMenuItemHref: (item: BreadcrumbMenuItem) => string = undefined;
  export let onSelectMenuItem: (item: string) => void = undefined;
  export let isCurrentPage = false;
</script>

<li class="flex items-center gap-x-2 p-2">
  <slot name="icon" />
  <div class="flex flex-row gap-x-1 items-center">
    <a
      {href}
      class={isCurrentPage
        ? "text-gray-800 font-medium"
        : "text-gray-500 hover:text-gray-600"}>{label}</a
    >
    {#if menuItems}
      <DropdownMenu.Root>
        <DropdownMenu.Trigger
          class="flex flex-col justify-center items-center transition-transform hover:translate-y-[2px] {isCurrentPage
            ? 'text-gray-800'
            : 'text-gray-500'}"
        >
          <CaretDownIcon size="14px" />
        </DropdownMenu.Trigger>
        <DropdownMenu.Content align="start" class="max-h-96 overflow-auto">
          {#each menuItems as item}
            <DropdownMenu.Item
              class="text-gray-800 hover:bg-gray-100 hover:text-gray-900 transition-colors"
              href={makeMenuItemHref ? makeMenuItemHref(item) : undefined}
              on:click={onSelectMenuItem
                ? () => onSelectMenuItem(item.key)
                : undefined}
            >
              {#if item.key === menuKey}
                <!-- If currently, selected show a check mark and bold the text -->
                <Check className="mr-2" />
                <span class="font-bold">{item.main}</span>
              {:else}
                <!-- If not selected, show an invisible check mark and normal text -->
                <Spacer className="mr-2" />
                <span>{item.main}</span>
              {/if}
            </DropdownMenu.Item>
          {/each}
        </DropdownMenu.Content>
      </DropdownMenu.Root>
    {/if}
  </div>
</li>
