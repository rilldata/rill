<script lang="ts" context="module">
  export type SubButtonKey = number | string;

  export type ButtonGroupContext = {
    registerSubButton?: (key: SubButtonKey) => void;
    toggleSubButton?: (key: SubButtonKey) => void;
    selectedKey?: Writable<SubButtonKey>;
    firstKey?: Writable<SubButtonKey>;
    lastKey?: Writable<SubButtonKey>;
    disabledKeys?: SubButtonKey[];
  };
  export const buttonGroupContext: ButtonGroupContext = {};
</script>

<script lang="ts">
  import { setContext, onDestroy } from "svelte";
  import { writable, get, Writable } from "svelte/store";
  import { createEventDispatcher } from "svelte";

  // If selectionRequired is true, then a sub button must be selected at all times.
  // In this case the button group behaves like a radio button.
  // If selectionRequired is false, It is possible to have no sub button selected.
  // If additionally, defaultKey is undefined, then no sub button is selected by default.
  export let selectionRequired = false;

  export let defaultKey: number | string | undefined;
  export let disabledKeys: (number | string)[] = [];

  const dispatch = createEventDispatcher();

  const subButtons = [];
  const selectedKey = writable(null);

  // if a default key is provided, then select it if it is not disabled
  // this Must be run reactively in case the default key changes or
  //  the set of disabled keys changes
  $: if (defaultKey !== undefined && !disabledKeys.includes(defaultKey)) {
    selectedKey.update(() => defaultKey);
  } else {
    selectedKey.update(() => null);
  }

  const firstKey = writable(null);
  const lastKey = writable(null);

  setContext(buttonGroupContext, {
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
      // if firstKey current value is null, then set it to the subButtonKey
      // being registered; otherwise, leave it as is
      firstKey.update((current) => current || subButtonKey);
      // always set lastKey to the subButtonKey being registered
      lastKey.update(() => subButtonKey);

      // if a selection is required,
      // and either no default key is provided or the default key is disabled,
      // and no sub button has yet been selected by the time this one is registered,
      // then the first sub button that is not disabled will be selected.
      // Note that if all sub buttons are disabled, then no sub button will be selected.
      if (
        selectionRequired &&
        (defaultKey === undefined || disabledKeys.includes(defaultKey)) &&
        get(selectedKey) === null &&
        !disabledKeys.includes(subButtonKey)
      ) {
        selectedKey.update(() => subButtonKey);
      }

      onDestroy(() => {
        const i = subButtons.indexOf(subButtonKey);
        subButtons.splice(i, 1);
        selectedKey.update((current) =>
          current === subButtonKey
            ? subButtons[i] || subButtons[subButtons.length - 1]
            : current
        );
      });
    },

    toggleSubButton: (subButton) => {
      // return if the sub button is disabled
      if (disabledKeys.includes(subButton)) return;

      const lastSelection = get(selectedKey);
      // if selection is required, then a sub button must always be selected
      // so if the sub button being toggled is already selected, then do nothing
      if (selectionRequired && lastSelection === subButton) return;

      // toggle the sub button: if it is selected, then deselect it;
      // otherwise, select it
      if (lastSelection === subButton) {
        selectedKey.set(null);
        dispatch("deselect-subbutton", subButton);
      } else {
        dispatch("deselect-subbutton", lastSelection);

        selectedKey.set(subButton);
        dispatch("select-subbutton", subButton);
      }
    },

    selectedKey,
    firstKey,
    lastKey,
    disabledKeys,
  });
</script>

<div
  class="flex flex-row w-fit rounded border border-gray-400 divide-x divide-gray-400"
>
  <slot />
</div>
