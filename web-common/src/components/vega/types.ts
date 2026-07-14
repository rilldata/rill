export type VLTooltipFormatter = (value: any) => string;

export type ExpressionFunction = Record<
  string,
  any | { fn: any; visitor?: any }
>;
