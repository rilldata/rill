export type ChipColors = {
  bgBaseClass: string;
  bgHoverClass: string;
  textClass: string;
  bgActiveClass: string;
  outlineClass: string;
};

export const defaultChipColors: ChipColors = {
  bgBaseClass: "bg-blue-50 dark:bg-blue-600",
  bgHoverClass: "hover:bg-blue-100 hover:dark:bg-blue-800",
  textClass: "text-blue-800 dark:text-blue-50",
  bgActiveClass: "bg-blue-100 dark:bg-blue-700",
  outlineClass: "outline-blue-400 dark:outline-blue-500",
};
