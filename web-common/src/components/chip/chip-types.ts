export type ChipColors = {
  bgBaseClass: string;
  bgHoverClass: string;
  bgActiveClass: string;
  outlineBaseClass: string;
  outlineHoverClass: string;
  outlineActiveClass: string;
  textClass: string;
};

/**
 * Use of !important for outlineActive to prevent flicker for hover and active states.
 */
export const defaultChipColors: ChipColors = {
  bgBaseClass: "bg-primary-50 dark:bg-primary-600",
  bgHoverClass: "hover:bg-primary-100 hover:dark:bg-primary-800",
  textClass: "text-primary-800 dark:text-primary-50",
  bgActiveClass: "bg-primary-100 dark:bg-primary-700",
  outlineBaseClass:
    "outline outline-1 outline-primary-100 dark:outline-primary-500",
  outlineHoverClass: "hover:outline-primary-200",
  outlineActiveClass: "!outline-primary-500 dark:outline-primary-500",
};

export const excludeChipColors: ChipColors = {
  bgBaseClass: "bg-gray-50 dark:bg-gray-700",
  bgHoverClass: "hover:bg-gray-100 hover:dark:bg-gray-600",
  textClass: "text-gray-600",
  bgActiveClass: "bg-gray-100 dark:bg-gray-600",
  outlineBaseClass: "outline outline-1 outline-gray-200",
  outlineHoverClass: "hover:outline-gray-300",
  outlineActiveClass: "!outline-gray-400",
};

// TODO: how does these colors fit into the new theme?
export const measureChipColors: ChipColors = {
  bgBaseClass: "bg-indigo-50 dark:bg-indigo-600",
  bgHoverClass: "hover:bg-indigo-100 hover:dark:bg-indigo-800",
  textClass: "text-indigo-800",
  bgActiveClass: "bg-indigo-100 dark:bg-indigo-600",
  outlineBaseClass:
    "outline outline-1 outline-indigo-200 dark:outline-indigo-500",
  outlineHoverClass: "hover:outline-indigo-300",
  outlineActiveClass: "!outline-indigo-500 dark:outline-indigo-500",
};
