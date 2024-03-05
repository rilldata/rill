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
  bgActiveClass: "bg-primary-100 dark:bg-primary-700",
  outlineBaseClass:
    "outline outline-1 outline-primary-100 dark:outline-primary-500",
  outlineHoverClass: "hover:outline-primary-200",
  outlineActiveClass: "!outline-primary-500 dark:outline-primary-500",
  textClass: "text-primary-800 dark:text-primary-50",
};

export const excludeChipColors: ChipColors = {
  bgBaseClass: "bg-gray-50 dark:bg-gray-700",
  bgHoverClass: "hover:bg-gray-100 hover:dark:bg-gray-600",
  bgActiveClass: "bg-gray-100 dark:bg-gray-600",
  outlineBaseClass: "outline outline-1 outline-gray-200",
  outlineHoverClass: "hover:outline-gray-300",
  outlineActiveClass: "!outline-gray-400",
  textClass: "text-gray-600",
};

export const measureChipColors: ChipColors = {
  bgBaseClass: "bg-secondary-50 dark:bg-secondary-600",
  bgHoverClass: "hover:bg-secondary-100 hover:dark:bg-secondary-800",
  bgActiveClass: "bg-secondary-100 dark:bg-secondary-600",
  outlineBaseClass:
    "outline outline-1 outline-secondary-200 dark:outline-secondary-500",
  outlineHoverClass: "hover:outline-secondary-300",
  outlineActiveClass: "!outline-secondary-500 dark:outline-secondary-500",
  textClass: "text-secondary-800",
};

export const timeChipColors: ChipColors = {
  bgBaseClass: "bg-white dark:bg-gray-700",
  bgHoverClass: "hover:bg-gray-100 hover:dark:bg-gray-600",
  bgActiveClass: "bg-gray-100 dark:bg-gray-600",
  outlineBaseClass: "outline outline-1 outline-gray-200",
  outlineHoverClass: "",
  outlineActiveClass: "",
  textClass: "text-gray-600",
};

export const specialChipColors: ChipColors = {
  bgBaseClass: "bg-purple-50 dark:bg-purple-600",
  bgHoverClass: "hover:bg-purple-100 hover:dark:bg-purple-800",
  bgActiveClass: "bg-purple-100 dark:bg-purple-600",
  outlineBaseClass:
    "outline outline-1 outline-purple-100 dark:outline-purple-500",
  outlineHoverClass: "hover:outline-purple-200",
  outlineActiveClass: "!outline-purple-500 dark:outline-purple-500",
  textClass: "text-purple-800",
};
