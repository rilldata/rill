<!-- @component
A simple menu of actions. When one is clicked, the callback fires,
and the menu closes.

This component is a fully-opinionated way of using selects.
A slot is provided to change the text within the button.

-->
<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import type { Alignment, Location } from "../../floating-element/types";
  import { SelectButton, WithSelectMenu } from "../index";
  import type { SelectMenuItem } from "../types";

  export let options: SelectMenuItem[];
  export let selection: SelectMenuItem;
  // this is fixed text that will always be displayed in the button
  export let fixedText = "";

  /* For multiSelect maintain array of keys in the consumer */
  export let multiSelect = false;
  export let location: Location = "bottom";
  export let alignment: Alignment = "start";
  export let distance = 16;
  export let active = false;

  export let paddingTop = 1;
  export let paddingBottom = 1;

  export let ariaLabel: undefined | string = undefined;

  const dispatch = createEventDispatcher();
</script>

<!-- wrap a WithSelectMenu with a SelectButton -->
<WithSelectMenu
  {paddingTop}
  {paddingBottom}
  {location}
  {alignment}
  {distance}
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
    on:click={toggleMenu}
    tailwindClasses="overflow-hidden"
    {active}
    label={ariaLabel}
  >
    {fixedText} <span class="font-bold truncate">{selection?.main}</span>
  </SelectButton>
</WithSelectMenu>
