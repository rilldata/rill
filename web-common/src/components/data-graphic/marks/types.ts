export interface Point {
  x: number;
  y: number;
  value: string;
  label: string;
  key: string;
  valueColorClass?: string;
  valueStyleClass?: string;
  labelColorClass?: string;
  labelStyleClass?: string;
  pointColorClass?: string;
  yOverride?: boolean;
  yOverrideLabel?: string;
  yOverrideStyleClass?: string;
}
