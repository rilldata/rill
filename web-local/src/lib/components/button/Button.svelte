<script lang="ts">
  import { createEventDispatcher } from "svelte";

  export let type: "primary" | "secondary" | "text";
  export let status: "info" | "error" = "info";
  export let disabled = false;
  export let compact = false;

  const dispatch = createEventDispatcher();

  const handleClick = (event: MouseEvent) => {
    if (!disabled) {
      dispatch("click", event);
    }
  };

  const disabledClasses = `disabled:cursor-not-allowed disabled:text-gray-700 disabled:bg-gray-200 disabled:border disabled:border-gray-400 disabled:opacity-50`;
  export const levels = {
    info: {
      primary: `bg-gray-800 border border-gray-800 hover:bg-gray-900 hover:border-gray-900 text-gray-100 hover:text-white focus:ring-blue-300`,
      secondary:
        "border border-gray-500 hover:bg-gray-200 hover:border-gray-200 focus:ring-blue-300",
      text: "text-gray-900 hover:bg-gray-300 focus:ring-blue-300",
    },
    error: {
      primary:
        "bg-red-200 border border-red-200 hover:bg-red-300 hover:border-red-300 text-red-800 active:ring-red-600 focus:ring-red-400",
      secondary:
        "border border-red-500 hover:bg-red-100 hover:border-red-600  focus:ring-red-400",
      text: "text-red-400 hover:bg-red-200  focus:ring-red-400",
    },
  };

  export function buttonClasses({
    /** one of thwee: primary, secondary, text */
    type = "primary",
    compact = false,
    status = "info",
    /** if you want to define a custom button style, use this string */
    customClasses = undefined,
  }) {
    return `
  ${
    compact ? "px-2 py-1" : "px-4 py-2"
  } rounded flex flex-row gap-x-2 items-center transition-transform duration-100
  focus:outline-none focus:ring-2
  ${customClasses ? customClasses : levels[status][type]}
  ${disabledClasses}
  `;
  }
</script>

<button
  style:height={compact ? "auto" : "36px"}
  {disabled}
  class={buttonClasses({ type, compact, status })}
  on:click={handleClick}
>
  <slot />
</button>
