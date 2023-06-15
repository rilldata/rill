<script lang="ts" context="module">
  export const buttonGroup = {};
</script>

<script lang="ts">
  import { setContext, onDestroy } from "svelte";
  import { writable, get } from "svelte/store";
  import { createEventDispatcher } from "svelte";

  // If selectionRequired is true, then a sub button must be selected at all times.
  // In this case the button group behaves like a radio button.
  // If selectionRequired is false, It is possible to have no sub button selected.
  // If additionally, defaultKey is undefined, then no sub button is selected by default.
  export let selectionRequired = false;

  export let defaultKey: number | string;
  export let disabledKeys: (number | string)[] = [];

  const dispatch = createEventDispatcher();

  const subButtons = [];
  const selectedSubButtonKey = writable(null);

  // if a default key is provided, then select it if it is not disabled
  // this Must be run reactively in case the default key changes or
  //  the set of disabled keys changes
  $: if (defaultKey !== undefined && !disabledKeys.includes(defaultKey)) {
    selectedSubButtonKey.update(() => defaultKey);
  } else {
    selectedSubButtonKey.update(() => null);
  }

  const firstSubButtonKey = writable(null);
  const lastSubButtonKey = writable(null);

  setContext(buttonGroup, {
    registerSubButton: (subButtonKey) => {
      if (
        typeof subButtonKey !== "number" &&
        typeof subButtonKey !== "string"
      ) {
        throw new Error(
          `Subbutton key must be a number or string. Received ${typeof subButtonKey}.`
        );
      }
      if (subButtons.includes(subButtonKey)) {
        throw new Error(
          `Subbutton with key ${subButtonKey} already registered. Subbutton keys must be unique.`
        );
      }
      subButtons.push(subButtonKey);
      // if firstSubButtonKey current value is null, then set it to the subButtonKey
      // being registered; otherwise, leave it as is
      firstSubButtonKey.update((current) => current || subButtonKey);
      // always set lastSubButtonKey to the subButtonKey being registered
      lastSubButtonKey.update(() => subButtonKey);

      // if a selection is required,
      // and either no default key is provided or the default key is disabled,
      // and no sub button has yet been selected by the time this one is registered,
      // then the first sub button that is not disabled will be selected.
      // Note that if all sub buttons are disabled, then no sub button will be selected.
      if (
        selectionRequired &&
        (defaultKey === undefined || disabledKeys.includes(defaultKey)) &&
        get(selectedSubButtonKey) === null &&
        !disabledKeys.includes(subButtonKey)
      ) {
        selectedSubButtonKey.update(() => subButtonKey);
      }

      onDestroy(() => {
        const i = subButtons.indexOf(subButtonKey);
        subButtons.splice(i, 1);
        selectedSubButtonKey.update((current) =>
          current === subButtonKey
            ? subButtons[i] || subButtons[subButtons.length - 1]
            : current
        );
      });
    },

    toggleSubButton: (subButton) => {
      // return if the sub button is disabled
      if (disabledKeys.includes(subButton)) return;

      const lastSelection = get(selectedSubButtonKey);
      // if selection is required, then a sub button must always be selected
      // so if the sub button being toggled is already selected, then do nothing
      if (selectionRequired && lastSelection === subButton) return;

      // toggle the sub button: if it is selected, then deselect it;
      // otherwise, select it
      if (lastSelection === subButton) {
        selectedSubButtonKey.set(null);
        dispatch("deselect-subbutton", subButton);
      } else {
        dispatch("deselect-subbutton", lastSelection);

        selectedSubButtonKey.set(subButton);
        dispatch("select-subbutton", subButton);
      }
    },

    selectedSubButton: selectedSubButtonKey,
    firstSubButtonKey,
    lastSubButtonKey,
    disabledKeys,
  });
</script>

<div
  class="flex flex-row w-fit rounded border border-gray-400 divide-x divide-gray-400"
>
  <slot />
</div>
