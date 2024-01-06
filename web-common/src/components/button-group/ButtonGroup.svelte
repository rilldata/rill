<script lang="ts" context="module">
  export type SubButtonKey = number | string;

  export type ButtonGroupContext = {
    registerSubButton?: (key: SubButtonKey) => void;
    subButtons?: Writable<SubButtonKey[]>;
    selectedKeys?: Writable<SubButtonKey[]>;
    disabledKeys?: Writable<SubButtonKey[]>;
    dispatch?: (type: "subbutton-click", detail: number | string) => boolean;
  };
  export const buttonGroupContext: ButtonGroupContext = {};
</script>

<script lang="ts">
  import { setContext, onDestroy } from "svelte";
  import { writable, get, Writable } from "svelte/store";
  import { createEventDispatcher } from "svelte";
  const dispatch = createEventDispatcher();

  export let selected: SubButtonKey[] = [];
  export let disabled: SubButtonKey[] = [];

  const subButtons: Writable<SubButtonKey[]> = writable([]);

  const selectedKeys: Writable<SubButtonKey[]> = writable([]);
  $: {
    selectedKeys.set(selected);
  }

  const disabledKeys: Writable<SubButtonKey[]> = writable([]);
  $: {
    disabledKeys.set(disabled);
  }

  setContext(buttonGroupContext, {
    registerSubButton: (subButtonKey) => {
      if (
        typeof subButtonKey !== "number" &&
        typeof subButtonKey !== "string"
      ) {
        throw new Error(
          `Subbutton value must be a number or string. Received ${typeof subButtonKey}.`,
        );
      }
      if (get(subButtons).includes(subButtonKey)) {
        throw new Error(
          `Subbutton with value ${subButtonKey} already registered. Subbutton values must be unique.`,
        );
      }
      subButtons.set([...get(subButtons), subButtonKey]);

      // called *during* initialization of sub button,
      // so applies to subbutton being registered
      onDestroy(() => {
        const i = get(subButtons).indexOf(subButtonKey);
        const newSubButtons = get(subButtons).slice();
        newSubButtons.splice(i, 1);
        subButtons.set(newSubButtons);

        selectedKeys.update((current) =>
          current.findIndex((key) => key === subButtonKey) === -1
            ? current
            : current.filter((key) => key !== subButtonKey),
        );
      });
    },
    subButtons,
    selectedKeys,
    disabledKeys,
    // Note: we pass the dispatch function here so that the subbutton
    // the subbutton can dispatch events "from" the parent button group.
    // Since the subbutton is slotted into the parent button group,
    // the wrapper div in the parent button group does not receive
    // the event normally and cannot forward it.
    dispatch,
  });
</script>

<div
  class="flex flex-row w-fit rounded border border-gray-300 divide-x divide-gray-300"
>
  <slot />
</div>
