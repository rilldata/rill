<script lang="ts">
  import { createEventDispatcher } from "svelte";

  export let type: "primary" | "secondary" | "highlighted" | "text" | "dashed" =
    "primary";
  export let status: "info" | "error" = "info";
  export let disabled = false;
  export let compact = false;
  export let submitForm = false;
  export let form = "";
  export let label: string | undefined = undefined;

  const dispatch = createEventDispatcher();

  const handleClick = (event: MouseEvent) => {
    if (!disabled) {
      dispatch("click", event);
    }
  };

  const disabledClasses = `disabled:cursor-not-allowed disabled:text-gray-700 disabled:bg-gray-200 disabled:border disabled:border-gray-400 disabled:opacity-50`;
  export const levels = {
    info: {
      primary: `bg-gray-800 text-white border rounded-sm border-gray-800 hover:bg-gray-700 hover:border-gray-700 focus:ring-blue-300`,
      secondary:
        "text-gray-800 border rounded-sm border-gray-300 shadow-sm hover:bg-gray-100 hover:text-gray-700 hover:border-gray-300 focus:ring-blue-300",
      highlighted:
        "text-gray-500 border border-gray-200 hover:bg-gray-200 hover:text-gray-600 hover:border-gray-200 focus:ring-blue-300 shadow-lg rounded-sm h-8 ",
      text: "text-gray-900 hover:bg-gray-300 focus:ring-blue-300",
      dashed:
        "text-gray-800 border border-dashed rounded-sm border-gray-300 shadow-sm hover:bg-gray-100 hover:text-gray-700 hover:border-gray-300 focus:ring-blue-300 ",
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
    /** one of four: primary, secondary, highlighted, text */
    type = "primary",
    compact = false,
    status = "info",
    /** if you want to define a custom button style, use this string */
    customClasses = undefined,
  }) {
    return `
  ${compact ? "px-2" : "px-3"} py-0.5 text-xs font-normal leading-snug
 flex flex-row gap-x-2 min-w-fit items-center justify-center transition-transform duration-100
  focus:outline-none focus:ring-2
  ${customClasses ? customClasses : levels[status][type]}
  ${disabledClasses}
  `;
  }

  const height = type === "highlighted" ? "32px" : compact ? "auto" : "28px";
</script>

<button
  style:height
  {disabled}
  class={buttonClasses({ type, compact, status })}
  on:click={handleClick}
  type={submitForm && "submit"}
  form={submitForm && form}
  aria-label={label}
>
  <slot />
</button>
