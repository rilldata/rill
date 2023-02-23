/** tailwind classes for elements in the workspace */

/** label / input container styling */
export const INPUT_ELEMENT_CONTAINER = {
  classes: "grid items-center gap-x-2",
  style: "grid-template-columns: 120px 224px",
};
export const CONFIG_TOP_LEVEL_INPUT_CONTAINER_CLASSES =
  "w-80 gap-x-4 flex items-center outline";

/** individual label class styling */
export const CONFIG_TOP_LEVEL_LABEL_CLASSES =
  "text-gray-500 font-medium w-[10em] text-[11px]";

/** active classes are used in selector items, where active-ness is determined with JS, not web APIs */
const activeSelectorClasses =
  "shadow-md outline-none ring-1 ring-gray-300 bg-white hover:bg-white";
const focusSelectorClasses =
  "focus:shadow-md focus:outline-none focus:ring-1 focus:ring-gray-300 focus:bg-white focus:hover:bg-white";
export const CONFIG_SELECTOR = {
  base: "overflow-hidden px-2 py-2 rounded bg-white border border-gray-200 hover:border-gray-300 hover:bg-gray-200 hover:text-gray-900 focus:outline-none focus:shadow-md",
  active: activeSelectorClasses,
  focus: focusSelectorClasses,
  error: "bg-red-50 border-red-400",
  distance: 8,
};

export const SELECTOR_BUTTON_TEXT_CLASSES = {
  selected: `font-semibold truncate`,
  unselected: `text-gray-600 truncate`,
};
export const SELECTOR_CONTAINER = {
  classes: "grow grid items-center",
  style: "grid-template-columns: 200px 24px",
};
