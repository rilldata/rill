interface DataProperties {
  metrics_view: string;
  filter?: string;
}

export interface TableProperties extends DataProperties {
  time_range: string;
  measures: string[];
  comparison_range?: string;
  row_dimensions?: string[];
  col_dimensions?: string[];
}
export interface TableTemplateT {
  table: TableProperties;
}

export interface MarkdownProperties {
  content: string;
  css?: { [key: string]: any };
}

export interface MarkdownTemplateT {
  markdown: MarkdownProperties;
}

export interface SelectProperties {
  valueField: string;
  labelField?: string;
  label?: string;
  tooltip?: string;
  placeholder?: string;
}

export interface SelectPropertiesT {
  select: SelectProperties;
}

export interface SwitchProperties {
  label: string;
  value: string;
  tooltip?: string;
}

export interface SwitchPropertiesT {
  switch: SwitchProperties;
}

export interface ImageProperties {
  url: string;
  css?: { [key: string]: any };
}

export interface ImageTemplateT {
  image: ImageProperties;
}

export type TemplateSpec =
  | TableTemplateT
  | MarkdownTemplateT
  | ImageTemplateT
  | SelectPropertiesT
  | SwitchPropertiesT;
