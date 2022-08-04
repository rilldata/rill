<!-- @component
A simple menu of actions. When one is clicked, the callback fires,
and the menu closes.

This component is a fully-opinionated way of using selects.
A slot is provided to change the text within the button.

-->
<script lang="ts">
  import { createEventDispatcher, setContext } from "svelte";

  import { SelectButton, WithSelectMenu } from "../";

  export let options;
  export let selection;

  export let tailwindClasses = undefined;
  export let activeTailwindClasses = undefined;

  /** TODO: figure out how we will support multiple selections
   * with the trigger. Until then, only allow single selections.
   */
  const multiple = false;

  /** When true, will make the trigger element a block-level element.
   * This is most useful when embedding a select menu in a table or wherever
   * a block-level treatment is needed.
   */
  export let block = false;

  export let level: undefined | "error" = undefined;

  export let dark: boolean = undefined;
  export let location: "left" | "right" | "top" | "bottom" = "bottom";
  export let alignment: "start" | "middle" | "end" = "start";
  export let distance = 16;

  export let active = false;

  if (dark) {
    setContext("rill:menu:dark", dark);
  }

  const dispatch = createEventDispatcher();
</script>

<!-- wrap a WithSelectMenu with a SelectButton -->
<WithSelectMenu
  {dark}
  {location}
  {alignment}
  {distance}
  on:select={(event) => {
    /** TODO: change this to work for multiple selections later. */
    selection = event.detail;
    dispatch("select", selection);
  }}
  bind:selection
  bind:options
  bind:active
  let:toggleMenu
  let:active
>
  <SelectButton
    on:click={toggleMenu}
    {activeTailwindClasses}
    {tailwindClasses}
    {active}
    {block}
    {level}
  >
    <slot>
      <div>
        {selection?.main || ""}
      </div>
    </slot>
  </SelectButton>
</WithSelectMenu>
