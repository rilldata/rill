<script lang="ts">
  import { MenuItem } from "@rgossiaux/svelte-headlessui";

  export let as: "a" | "button" = "a";
  export let disabled = false;

  function classNames(...classes: (string | false | null | undefined)[]) {
    return classes.filter(Boolean).join(" ");
  }

  function resolveClass({
    active,
    disabled,
  }: {
    active: boolean;
    disabled: boolean;
  }) {
    return classNames(
      "flex items-center gap-x-2 px-3 py-1.5 text-xs font-normal leading-none w-full whitespace-nowrap",
      active ? "bg-gray-100 text-gray-900 cursor-pointer" : "text-gray-700",
      disabled && "cursor-not-allowed opacity-50",
    );
  }
</script>

<MenuItem {as} {disabled} on:click class={resolveClass}>
  <slot />
</MenuItem>
