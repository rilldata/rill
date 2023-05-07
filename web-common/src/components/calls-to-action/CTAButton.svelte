<script lang="ts">
  import { createEventDispatcher } from "svelte";

  export let variant: "primary" | "primary-outline" | "secondary" = "primary";
  export let disabled = false;

  function getVariantClass(variant: string) {
    console.log;
    switch (variant) {
      case "primary":
        return "border-blue-600 bg-blue-600 text-white hover:bg-blue-500 hover:border-blue-500";
      case "primary-outline":
        return "border-blue-300 text-blue-600 hover:bg-slate-100 hover:border-gray-100";
      case "secondary":
        return "text-slate-600 border-slate-300 hover:bg-slate-100";
    }
  }

  const dispatch = createEventDispatcher();

  const handleClick = (event: MouseEvent) => {
    if (!disabled) {
      dispatch("click", event);
    }
  };

  const disabledClasses = `disabled:cursor-not-allowed disabled:text-gray-700 disabled:bg-gray-200 disabled:border disabled:border-gray-400 disabled:opacity-50`;
</script>

<button
  class="text-sm w-full max-w-[400px] h-10 border rounded-sm {getVariantClass(
    variant
  )} {disabled && disabledClasses}"
  on:click={handleClick}
  {disabled}
>
  <slot />
</button>
