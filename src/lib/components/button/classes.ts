export const levels = {
  primary:
    "bg-gray-800 border border-gray-800 hover:bg-gray-900 hover:border-gray-900 text-gray-100 hover:text-white",
  secondary: "border border-gray-500 hover:bg-gray-200 hover:border-gray-200",
  text: "text-gray-900 hover:bg-gray-300",
};

export function buttonClasses({
  /** one of thwee: primary, secondary, text */
  type = "primary",
  compact = false,
  /** if you want to define a custom button style, use this string */
  customClasses = undefined,
}) {
  return `
  ${
    compact ? "px-2 py-1" : "px-4 py-2"
  } rounded flex flex-row gap-x-2 items-center transition-transform duration-100
  ${customClasses ? customClasses : levels[type]}
  disabled:cursor-not-allowed disabled:text-gray-700 disabled:bg-gray-200 disabled:border disabled:border-gray-400 disabled:opacity-50
  `;
}
