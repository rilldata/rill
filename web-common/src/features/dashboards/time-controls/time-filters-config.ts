export type TimeFiltersConfig = {
  hidePan?: boolean;
  showTimeDimensionSelector?: boolean;
  showComparisonSelector?: boolean;
  allowCustomTimeRange?: boolean;
  skipDefaultTimeRange?: boolean;
  skipTimeGrain?: boolean;
  lockTimeZone?: boolean;
  showFullRange?: boolean;
  showWatermark?: boolean;
  showInheritRange?: boolean;

  log?: boolean;

  side?: "top" | "right" | "bottom" | "left";
};
