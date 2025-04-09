export const MIN_DIMENSION_COLUMN_WIDTH = 150;
export const LEADERBOARD_WRAPPER_PADDING = 56;

export function getDimensionColumnWidth(
  wrapperWidth: number,
  contextColWidth: number,
  measureNames: string[],
) {
  if (!wrapperWidth) {
    return MIN_DIMENSION_COLUMN_WIDTH;
  }
  return Math.max(
    MIN_DIMENSION_COLUMN_WIDTH,
    wrapperWidth -
      contextColWidth * measureNames.length -
      LEADERBOARD_WRAPPER_PADDING,
  );
}
