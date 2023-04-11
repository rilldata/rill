<script lang="ts">
  import { IconButton } from "@rilldata/web-common/components/button";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import SimpleActionMenu from "@rilldata/web-common/components/menu/wrappers/SimpleActionMenu.svelte";

  export let label: string;
  export let isActive = false;
  export let options: { main: string; callback: () => void }[] = [];

  const inactiveClass = "text-gray-500";
  const activeClass = "text-gray-800 font-semibold";
</script>

<li class={isActive ? activeClass : inactiveClass}>
  <div class="flex flex items-center gap-x-3 p-2 ">
    <slot name="icon" />
    <div class="flex items-center gap-x-1">
      <span>
        {label}
      </span>
      {#if options}
        <SimpleActionMenu {options} let:toggleMenu minWidth="0px" distance={4}>
          <IconButton on:click={toggleMenu}>
            <CaretDownIcon size="14px" className="text-gray-500" />
          </IconButton>
        </SimpleActionMenu>
      {/if}
    </div>
  </div>
</li>
