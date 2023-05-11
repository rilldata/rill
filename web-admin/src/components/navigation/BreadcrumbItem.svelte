<script lang="ts">
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import WithSelectMenu from "@rilldata/web-common/components/menu/wrappers/WithSelectMenu.svelte";

  export let label: string;
  export let href: string;
  export let menuOptions: { key: string; main: string }[] = [];
  export let menuKey: string;
  export let onSelectMenuOption: (option: string) => void = undefined;
  export let isCurrentPage = false;
</script>

<li class="flex flex items-center gap-x-2 p-2">
  <slot name="icon" />
  <div class="flex flex-row gap-x-1 items-center">
    <a
      {href}
      class={isCurrentPage
        ? "text-gray-800 font-medium"
        : "text-gray-500 hover:text-gray-600"}>{label}</a
    >
    {#if menuOptions}
      <WithSelectMenu
        minWidth="0px"
        distance={4}
        options={menuOptions}
        selection={{
          key: menuKey,
          main: label,
        }}
        on:select={({ detail: { key } }) => onSelectMenuOption(key)}
        let:toggleMenu
      >
        <button
          class="flex flex-col justify-center items-center transition-transform hover:translate-y-[2px] {isCurrentPage
            ? 'text-gray-800'
            : 'text-gray-500'}"
          on:click={toggleMenu}
        >
          <CaretDownIcon size="14px" />
        </button>
      </WithSelectMenu>
    {/if}
  </div>
</li>
