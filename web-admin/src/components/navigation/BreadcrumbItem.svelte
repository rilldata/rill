<script lang="ts">
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import WithSelectMenu from "@rilldata/web-common/components/menu/wrappers/WithSelectMenu.svelte";

  export let label: string;
  export let isCurrentPage = false;
  export let menuOptions: { key: string; main: string }[] = [];
  export let onSelectMenuOption: (option: string) => void = undefined;

  const activeClass = "text-gray-800 font-semibold";
  const inactiveClass = "text-gray-500";

  let hovered = false;
  function setHovered(value: boolean) {
    hovered = value;
  }
</script>

<li class="flex flex items-center gap-x-1 p-2">
  <slot name="icon" />
  {#if !menuOptions}
    <span class={isCurrentPage ? activeClass : inactiveClass}>{label}</span>
  {:else}
    <WithSelectMenu
      minWidth="0px"
      distance={4}
      options={menuOptions}
      selection={{
        key: label,
        main: label,
      }}
      on:select={({ detail: { key } }) => onSelectMenuOption(key)}
      let:toggleMenu
    >
      <button
        class="flex flex-row gap-x-1 items-center {isCurrentPage
          ? activeClass
          : inactiveClass}"
        on:click={toggleMenu}
        on:mouseenter={() => setHovered(true)}
        on:mouseleave={() => setHovered(false)}
      >
        <span>{label}</span>
        <div class="transition-transform" class:translate-y-2={hovered}>
          <CaretDownIcon size="14px" />
        </div>
      </button>
    </WithSelectMenu>
  {/if}
</li>
