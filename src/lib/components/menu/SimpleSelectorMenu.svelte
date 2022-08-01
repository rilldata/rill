<!-- @component
A simple menu of actions. When one is clicked, the callback fires,
and the menu closes.

-->
<script lang="ts">
  import { createEventDispatcher, setContext } from "svelte";
  import CaretDownIcon from "../icons/CaretDownIcon.svelte";
  import Check from "../icons/Check.svelte";
  import CheckBox from "../icons/CheckBox.svelte";
  import CheckCircle from "../icons/CheckCircle.svelte";
  import EmptyBox from "../icons/EmptyBox.svelte";
  import EmptyCircle from "../icons/EmptyCircle.svelte";
  import Spacer from "../icons/Spacer.svelte";

  import Menu from "./Menu.svelte";
  import MenuItem from "./MenuItem.svelte";
  import WithFloatingMenu from "./WithFloatingMenu.svelte";

  export let options;
  export let selections = [];
  export let style = "obvious";
  export let multiple = false;

  export let tailwindClasses = undefined;
  export let activeTailwindClasses = undefined;

  /** When true, will make the trigger element a block-level element.
   */
  export let block = false;

  export let dark: boolean = undefined;
  export let location: "left" | "right" | "top" | "bottom" = "bottom";
  export let alignment: "start" | "middle" | "end" = "start";
  export let distance = 16;

  export let active = false;

  if (dark) {
    setContext("rill:menu:dark", dark);
  }

  const dispatch = createEventDispatcher();

  let temporarilySelectedKey;
  function createOnClickHandler(
    main,
    right,
    description,
    key,
    index,
    closeEventHandler
  ) {
    return async () => {
      // single-select: do nothing if already selected
      if (!multiple && isSelected(selections, key)) {
        return;
      }
      // set temporarily selected to get the icon to change instantly, then wait for tick
      // proceed with rest of update
      if (multiple) {
        // check to see if exists
        // if not, add.
        if (isSelected(selections, key)) {
          selections = [...selections.filter((s) => s.key !== key)];
        } else {
          selections = [
            ...selections,
            { main, right, key, description, index },
          ];
        }
      } else {
        // replace selected with a single value.
        selections = [{ main, right, description, key, index }];
      }
      dispatch("select", selections);
      if (!multiple) closeEventHandler();

      temporarilySelectedKey = undefined;
    };
  }

  function isSelected(selections, key) {
    return selections?.some((selection) => selection.key === key);
  }
</script>

<WithFloatingMenu
  bind:active
  {location}
  {alignment}
  {distance}
  let:handleClose
  let:toggleMenu
>
  <button
    class="
    {block ? 'flex w-full h-full px-2' : 'inline-flex w-max rounded px-1'} 
      items-center gap-x-2 justify-between {!active
      ? 'hover:bg-gray-100'
      : `${activeTailwindClasses} bg-gray-200`}
      {tailwindClasses}"
    on:click={toggleMenu}
  >
    <slot>
      <div>
        {selections[0].main}
      </div>
    </slot>
    <CaretDownIcon />
  </button>

  <Menu
    slot="menu"
    {dark}
    on:lose-focus={() => {
      if (active) handleClose();
    }}
    on:escape={handleClose}
  >
    {#each options as { key, main, description, right }, i}
      {@const selected = isSelected(selections, key)}
      <MenuItem
        icon
        animateSelect={!multiple}
        on:before-select={() => {
          temporarilySelectedKey = key;
        }}
        on:select={createOnClickHandler(
          main,
          right,
          description,
          key,
          i,
          handleClose
        )}
        {selected}
      >
        <svelte:fragment slot="icon">
          <!-- this conditional will make the circle check appear briefly before the menu closes
          in the case of a single-select menu. -->
          {#if !multiple}
            {#if (temporarilySelectedKey !== undefined && temporarilySelectedKey === key) || (temporarilySelectedKey === undefined && selected)}
              {#if style === "obvious"}<CheckCircle />{:else}<Check />{/if}
            {:else if style === "obvious"}
              <EmptyCircle />{:else}<Spacer />
            {/if}
            <!-- multi -->
          {:else if selected}
            {#if style === "obvious"}<CheckBox />{:else}<Check />{/if}
          {:else if style === "obvious"}<EmptyBox />{:else}<Spacer />{/if}
        </svelte:fragment>

        {main}
        <svelte:fragment slot="description">
          {#if description}
            {description}
          {/if}
        </svelte:fragment>
        <svelte:fragment slot="right">
          {right || ""}
        </svelte:fragment>
      </MenuItem>
    {/each}
  </Menu>
</WithFloatingMenu>
