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
  "text-fg-secondary font-medium w-[10em] text-[11px]";

/** active classes are used in selector items, where active-ness is determined with JS, not web APIs */
const activeSelectorClasses =
  "shadow-md outline-none ring-1 ring-gray-300 bg-surface-subtle hover:bg-surface-subtle";
const focusSelectorClasses =
  "focus:shadow-md focus:outline-none focus:ring-1 focus:ring-gray-300 focus:bg-surface-subtle focus:hover:bg-surface-subtle";
export const CONFIG_SELECTOR = {
  base: "overflow-hidden px-2 py-2 rounded border focus:outline-none focus:shadow-md",
  active: activeSelectorClasses,
  focus: focusSelectorClasses,
  info: "bg-surface-subtle   hover:border-gray-300 hover:bg-surface-hover hover:text-fg-primary",
  // FIXME: we should probably redo how we're handling these styles.
  error:
    "bg-red-100 border-red-500 text-red-800 hover:bg-red-200 hover:border-red-600",
  activeError:
    "shadow-md outline-none ring-1 ring-red-400 bg-red-100 hover:bg-red-200",
  distance: 8,
};

export const SELECTOR_BUTTON_TEXT_CLASSES = {
  selected: `font-semibold truncate`,
  unselected: `text-fg-secondary truncate`,
};
export const SELECTOR_CONTAINER = {
  classes: "grow grid items-center",
  style: "grid-template-columns: 200px 24px",
};
