type NativeInputTypes = "text" | "number" | "boolean" | "textArea";
type SemanticInputTypes =
  | "metrics_view"
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

export interface ComponentInputParam {
  type: InputType;
  label?: string;
  showInUI?: boolean; // If not specified, can assume true
  required?: boolean; // If not specified, can assume true
  description?: string; // Tooltip description for the input
}
