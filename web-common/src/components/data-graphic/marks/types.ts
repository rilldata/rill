export interface Point {
  x: number;
  y: number;
  value?: string;
  label: string;
  key: string;
  valueColorClass?: string;
  valueStyleClass?: string;
  labelColorClass?: string;
  labelStyleClass?: string;
  pointColor?: string;
  pointOpacity?: number;
  yOverride?: boolean;
  yOverrideLabel?: string;
  yOverrideStyleClass?: string;
}

export interface YValue {
  y: string | number | Date | undefined | null;
  name?: string | null;
  color?: string;
  isTimeComparison?: boolean;
}
