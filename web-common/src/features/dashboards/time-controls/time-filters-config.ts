export type TimeFiltersConfig = {
  hidePan?: boolean;
  showGrainSelector?: boolean; // TODO: not used in new picker?
  showTimeDimensionSelector?: boolean;
  showComparisonSelector?: boolean;
  allowCustomTimeRange?: boolean;
  lockTimeZone?: boolean;
  showFullRange?: boolean;
  showWatermark?: boolean;

  side?: "top" | "right" | "bottom" | "left";
};
