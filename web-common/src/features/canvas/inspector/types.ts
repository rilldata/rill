import type { ComponentAlignment } from "@rilldata/web-common/features/canvas/components/types";
import type {
  ChartLegend,
  ChartSortDirectionOptions,
} from "@rilldata/web-common/features/components/charts/types";

type NativeInputTypes = "text" | "number" | "boolean" | "textArea" | "select";
type SemanticInputTypes = "metrics" | "measure" | "dimension" | "multi_fields";
type ChartInputTypes = "positional" | "mark" | "tooltip" | "config";
type CustomInputTypes =
  | "rill_time"
  | "sparkline"
  | "comparison_options"
  | "switcher_tab";
type PositionalInputTypes = "alignment";

export type InputType =
  | NativeInputTypes
  | SemanticInputTypes
  | ChartInputTypes
  | CustomInputTypes
  | PositionalInputTypes;

export type FilterInputTypes = "time_filters" | "dimension_filters";

export type FieldType = "measure" | "dimension" | "time";

export type SortSelectorConfig = {
  enable: boolean;
  customSortItems?: string[];
  defaultSort?: string;
  options?: ChartSortDirectionOptions[];
};

export type ChartFieldInput = {
  type: FieldType | "value";
  excludedValues?: string[];
  axisTitleSelector?: boolean;
  hideTimeDimension?: boolean;
  originSelector?: boolean;
  sortSelector?: SortSelectorConfig;
  limitSelector?: { defaultLimit: number };
  colorMappingSelector?: {
    enable: boolean;
    values?: string[];
    isContinuous?: boolean;
  };
  colorRangeSelector?: {
    enable: boolean;
  };
  nullSelector?: boolean;
  labelAngleSelector?: boolean;
  axisRangeSelector?: boolean;
  multiFieldSelector?: boolean;
  /**
   * For combo charts individual field can be a bar or line chart.
   */
  markTypeSelector?: boolean;
  /**
   * The default legend position for the chart.
   * If this key is not specified, legend selector will not be shown.
   */
  defaultLegendOrientation?: ChartLegend;
  /**
   * For measures toggle for displaying measure total value
   */
  totalSelector?: boolean;
};

export interface ComponentInputParam {
  type: InputType;
  label?: string;
  showInUI?: boolean; // If not specified, can assume true
  optional?: boolean;
  description?: string; // Tooltip description for the input
  meta?: {
    allowedTypes?: FieldType[]; // Specify which field types are allowed for multi-field selection
    defaultAlignment?: ComponentAlignment;
    chartFieldInput?: ChartFieldInput;
    layout?: "default" | "grouped";
    /**
     * If true, the boolean input will be inverted. This is useful when true
     * is the intended default state
     */
    invertBoolean?: boolean;
    [key: string]: any;
  };
}

export interface FilterInputParam {
  type: FilterInputTypes;
  meta?: Record<string, any>;
}

export type AllKeys<T> = T extends any ? keyof T : never;

export interface InputParams<T> {
  options: Partial<Record<AllKeys<T>, ComponentInputParam>>;
  filter: Partial<Record<FilterInputTypes, FilterInputParam>> | [];
}
