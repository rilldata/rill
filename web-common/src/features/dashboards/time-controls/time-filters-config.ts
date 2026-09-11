export type TimeFiltersConfig = {
  hidePan?: boolean;
  canPanLeft?: boolean;
  canPanRight?: boolean;

  showTimeDimensionSelector?: boolean;
  lockTimeZone?: boolean;
  showFullRange?: boolean;
  showWatermark?: boolean;

  side?: "top" | "right" | "bottom" | "left";
};
