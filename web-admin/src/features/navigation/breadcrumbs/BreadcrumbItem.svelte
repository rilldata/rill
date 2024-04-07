<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
  import Spacer from "@rilldata/web-common/components/icons/Spacer.svelte";

  export let label: string;
  export let href: string;
  export let menuOptions: { key: string; main: string }[] = [];
  export let menuKey: string;
  export let onSelectMenuOption: (option: string) => void = undefined;
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
    {#if menuOptions}
      <DropdownMenu.Root>
        <DropdownMenu.Trigger
          class="flex flex-col justify-center items-center transition-transform hover:translate-y-[2px] {isCurrentPage
            ? 'text-gray-800'
            : 'text-gray-500'}"
        >
          <CaretDownIcon size="14px" />
        </DropdownMenu.Trigger>
        <DropdownMenu.Content align="start" class="max-h-96 overflow-auto">
          {#each menuOptions as option}
            <DropdownMenu.Item on:click={() => onSelectMenuOption(option.key)}>
              {#if option.key === menuKey}
                <!-- If currently, selected show a check mark and bold the text -->
                <Check className="mr-2" />
                <span class="font-bold">{option.main}</span>
              {:else}
                <!-- If not selected, show an invisible check mark and normal text -->
                <Spacer className="mr-2" />
                <span>{option.main}</span>
              {/if}
            </DropdownMenu.Item>
          {/each}
        </DropdownMenu.Content>
      </DropdownMenu.Root>
    {/if}
  </div>
</li>
