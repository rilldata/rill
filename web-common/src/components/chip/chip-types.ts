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

// TODO: how does these colors fit into the new theme?
export const measureChipColors = {
  bgBaseClass: "bg-secondary-50 dark:bg-secondary-600",
  bgHoverClass: "hover:bg-secondary-100 hover:dark:bg-secondary-800",
  textClass: "text-secondary-800",
  bgActiveClass: "bg-secondary-100 dark:bg-secondary-600",
  outlineClass:
    "outline outline-1 outline-secondary-200 dark:outline-secondary-500 hover:outline-secondary-200",
  outlineActiveClass: "!outline-secondary-500 dark:outline-secondary-500",
};

export const timeChipColors = {
  bgBaseClass: "bg-gray-50 dark:bg-gray-700",
  bgHoverClass: "hover:bg-gray-100 hover:dark:bg-gray-600",
  textClass: "text-gray-600",
  bgActiveClass: "bg-gray-100 dark:bg-gray-600",
  outlineClass: "",
  outlineActiveClass: "",
};

export const specialChipColors = {
  bgBaseClass: "bg-purple-50 dark:bg-purple-600",
  bgHoverClass: "hover:bg-purple-100 hover:dark:bg-purple-800",
  textClass: "text-purple-800",
  bgActiveClass: "bg-purple-100 dark:bg-purple-600",
  outlineClass:
    "outline outline-1 outline-purple-100 dark:outline-purple-500 hover:outline-purple-200",
  outlineActiveClass: "!outline-purple-500 dark:outline-purple-500",
};
