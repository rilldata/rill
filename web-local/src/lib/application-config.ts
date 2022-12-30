/** Application-level constants and shared settings
 * ------------------------------------------------
 * spacing, tween lengths, etc.
 */

import { cubicOut as easing } from "svelte/easing";

export const DEFAULT_INSPECTOR_WIDTH = 360;
export const DEFAULT_NAV_WIDTH = 240;
export const DEFAULT_PREVIEW_TABLE_HEIGHT = 400;

/** parameters used in the column profile view & elsewhere */
export const COLUMN_PROFILE_CONFIG = {
  /** The null percentage should be _just_ big enough to show x 100.0%
   * For MD IO 0.4, this is 74px.
   */
  nullPercentageWidth: 44,
  mediumCutoff: 300,
  compactBreakpoint: 300,
  hideRight: 0,
  hideNullPercentage: 0,
  summaryVizWidth: { medium: 68, small: 64 },
  exampleWidth: { medium: 204, small: 132 },
  fontSize: 12,
};

export const TOOLTIP_STRING_LIMIT = 200;

export function collapseInspectorCTAButton(width) {
  return !(width < 398);
}

/** layout constants  */
export const SIDE_PAD = 0;
export const SURFACE_SLIDE_DURATION = 400;
export const LIST_SLIDE_DURATION = 200;
export const SURFACE_SLIDE_EASING = easing;
export const SURFACE_DRAG_DURATION = 50;

/** level color tokens */
export const level = {
  info: {
    text: "text-blue-800",
    border: "border-blue-600",
    bg: "bg-blue-50",
    bgSecondary: "bg-blue-50",
    borderSecondary: "border-blue-500",
  },
  warning: {
    text: "text-yellow-800",
    border: " border-yellow-600",
    bg: "bg-yellow-100",
    bgSecondary: "bg-yellow-50",
    borderSecondary: "border-yellow-400",
  },
  error: {
    text: "text-red-800",
    border: "border-red-600",
    bg: "bg-red-100",
    borderSecondary: "border-red-500",
    bgSecondary: "bg-red-50",
  },
};
