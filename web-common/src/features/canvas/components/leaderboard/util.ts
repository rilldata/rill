export const MIN_DIMENSION_COLUMN_WIDTH = 150;
export const DEFAULT_DIMENSION_COLUMN_WIDTH = 164;
export const LEADERBOARD_WRAPPER_PADDING = 56;

export function getDimensionColumnWidth(
  wrapperWidth: number,
  contextColWidth: number,
) {
  if (!wrapperWidth) {
    return DEFAULT_DIMENSION_COLUMN_WIDTH;
  }
  return Math.max(
    MIN_DIMENSION_COLUMN_WIDTH,
    wrapperWidth - contextColWidth - LEADERBOARD_WRAPPER_PADDING,
  );
}
