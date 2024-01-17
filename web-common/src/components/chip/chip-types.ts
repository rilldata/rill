export type ChipColors = {
  bgBaseClass: string;
  bgHoverClass: string;
  textClass: string;
  bgActiveClass: string;
  outlineClass: string;
  outlineActiveClass: string;
};

/**
 * Use of !important for outlineActive to prevent flicker for hover and active states.
 */
export const defaultChipColors: ChipColors = {
  bgBaseClass: "bg-primary-50 dark:bg-primary-600",
  bgHoverClass: "hover:bg-primary-100 hover:dark:bg-primary-800",
  textClass: "text-primary-800 dark:text-primary-50",
  bgActiveClass: "bg-primary-100 dark:bg-primary-700",
  outlineClass:
    "outline outline-1 outline-primary-100 dark:outline-primary-500 hover:outline-primary-200",
  outlineActiveClass: "!outline-primary-500 dark:outline-primary-500",
};

export const excludeChipColors = {
  bgBaseClass: "bg-gray-50 dark:bg-gray-700",
  bgHoverClass: "hover:bg-gray-100 hover:dark:bg-gray-600",
  textClass: "text-gray-600",
  bgActiveClass: "bg-gray-100 dark:bg-gray-600",
  outlineClass: "outline outline-1 outline-gray-200",
  outlineActiveClass: "!outline-gray-400",
};
