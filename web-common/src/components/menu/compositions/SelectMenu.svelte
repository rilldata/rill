<!-- @component
A simple menu of actions. When one is clicked, the callback fires,
and the menu closes.

This component is a fully-opinionated way of using selects.
A slot is provided to change the text within the button.

-->
<script lang="ts">
  import { createEventDispatcher, setContext } from "svelte";
  import type { Alignment, Location } from "../../floating-element/types";
  import { SelectButton, WithSelectMenu } from "../index";

  export let options;
  export let selection;
  export let tailwindClasses = undefined;
  export let activeTailwindClasses = undefined;
  /** When true, will make the trigger element a block-level element.
   * This is most useful when embedding a select menu in a table or wherever
   * a block-level treatment is needed.
   */
  export let block = false;
  export let level: undefined | "error" = undefined;
  export let dark: boolean = undefined;
  export let disabled = false;

  /* For multiSelect maintain array of keys in the consumer */
  export let multiSelect = false;
  export let location: Location = "bottom";
  export let alignment: Alignment = "start";
  export let distance = 16;
  export let active = false;

  export let paddingTop: number = 1;
  export let paddingBottom: number = 1;

  if (dark) {
    setContext("rill:menu:dark", dark);
  }

  const dispatch = createEventDispatcher();
</script>

<!-- wrap a WithSelectMenu with a SelectButton -->
<WithSelectMenu
  {paddingTop}
  {paddingBottom}
  {dark}
  {location}
  {alignment}
  {distance}
  {disabled}
  {multiSelect}
  on:select={(event) => {
    if (!multiSelect) selection = event.detail;
    dispatch("select", selection);
  }}
  bind:selection
  bind:options
  bind:active
  let:toggleMenu
  let:active
>
  <SelectButton
    on:click={() => {
      if (!disabled) {
        console.log("hmm");
        toggleMenu();
      }
    }}
    {tailwindClasses}
    {activeTailwindClasses}
    {active}
    {block}
    {disabled}
    {level}
  >
    <slot>
      <div>
        {selection?.main || ""}
      </div>
    </slot>
  </SelectButton>
</WithSelectMenu>
