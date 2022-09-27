/** Application-level constants and shared settings
 * ------------------------------------------------
 * spacing, tween lengths, etc.
 */

/** parameters used in the column profile view & elsewhere */
export const COLUMN_PROFILE_CONFIG = {
  /** The null percentage should be _just_ big enough to show x 100.0%
   * For MD IO 0.4, this is 74px.
   */
  nullPercentageWidth: 74,
  mediumCutoff: 300,
  compactBreakpoint: 350,
  hideRight: 325,
  hideNullPercentage: 399,
  summaryVizWidth: { medium: 84, small: 60 },
  exampleWidth: { medium: 204, small: 132 },
};

export function collapseInspectorCTAButton(width) {
  return !(width < 398);
}
