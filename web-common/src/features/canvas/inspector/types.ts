type NativeInputTypes = "text" | "number" | "boolean" | "textArea";
type SemanticInputTypes =
  | "metrics"
  | "measure"
  | "dimension"
  | "multi_measures"
  | "multi_dimensions";
type ChartInputTypes = "positional" | "mark" | "tooltip";
type CustomInputTypes = "rill_time";

export type InputType =
  | NativeInputTypes
  | SemanticInputTypes
  | ChartInputTypes
  | CustomInputTypes;

export type FilterInputTypes =
  | "time_range"
  | "comparison_range"
  | "grain"
  | "dimension_filters";

export interface ComponentInputParam {
  type: InputType;
  label?: string;
  showInUI?: boolean; // If not specified, can assume true
  optional?: boolean;
  description?: string; // Tooltip description for the input
  meta?: Record<string, string>; // Any additional metadata
}

export interface FilterInputParam {
  type: FilterInputTypes;
  label: string;
}

export interface InputParams<T> {
  component: Partial<Record<keyof T, ComponentInputParam>>;
  filter: Partial<Record<FilterInputTypes, FilterInputParam>> | [];
}
