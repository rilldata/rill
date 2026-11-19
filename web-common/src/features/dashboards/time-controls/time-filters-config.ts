export type TimeFiltersConfig = {
  hidePan?: boolean;
  showTimeDimensionSelector?: boolean;
  showComparisonSelector?: boolean;
  allowCustomTimeRange?: boolean;
  skipDefaultTimeRange?: boolean;
  skipTimeGrain?: boolean;
  showFullRange?: boolean;
  showWatermark?: boolean;
  showInheritRange?: boolean;
  showGrainPicker?: boolean;

  side?: "top" | "right" | "bottom" | "left";
};
