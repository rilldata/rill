export type TimeFiltersConfig = {
  hidePan?: boolean;
  canPanLeft?: boolean;
  canPanRight?: boolean;

  showTimeDimensionSelector?: boolean;
  allowCustomTimeRange?: boolean;
  showDefaultItem: boolean;
  lockTimeZone?: boolean;
  showFullRange?: boolean;
  showWatermark?: boolean;

  side: "top" | "right" | "bottom" | "left";
};
