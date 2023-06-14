<script lang="ts" context="module">
  export const buttonGroup = {};
</script>

<script lang="ts">
  import { setContext, onDestroy } from "svelte";
  import { writable } from "svelte/store";

  // If selectionRequired is true, then a sub button must be selected at all times.
  // In this case the button group behaves like a radio button.
  // If selectionRequired is false, It is possible to have no sub button selected.
  // If additionally, defaultKey is undefined, then no sub button is selected by default.
  export let selectionRequired: boolean;

  export let defaultKey: number | string;

  const subButtons = [];
  const selectedSubButtonKey = writable(defaultKey);

  const firstSubButtonKey = writable(null);
  const lastSubButtonKey = writable(null);

  setContext(buttonGroup, {
    registerSubButton: (subButtonKey) => {
      subButtons.push(subButtonKey);
      // if firstSubButtonKey current value is null, then set it to the subButtonKey
      // being registered; otherwise, leave it as is
      firstSubButtonKey.update((current) => current || subButtonKey);
      // always set lastSubButtonKey to the subButtonKey being registered
      lastSubButtonKey.update(() => subButtonKey);

      // If no default is provided, and a selection is required, then when the first
      // sub button is registered, it will be selected
      if (selectionRequired && defaultKey === undefined) {
        selectedSubButtonKey.update((current) => current || subButtonKey);
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

    selectSubButton: (subButton) => {
      selectedSubButtonKey.set(subButton);
    },

    selectedSubButton: selectedSubButtonKey,
    firstSubButtonKey,
    lastSubButtonKey,
  });
</script>

<div class="flex flex-row">
  <slot />
</div>
