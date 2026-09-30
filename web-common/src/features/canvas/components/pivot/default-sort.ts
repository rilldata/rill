import {
  getTimeGrainFromDimension,
  isTimeDimension,
} from "@rilldata/web-common/features/dashboards/pivot/pivot-utils";
import {
  COMPARISON_DELTA,
  COMPARISON_PERCENT,
  COMPARISON_VALUE,
  type PivotDataStoreConfig,
  type PivotSortTarget,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
import { convertISOStringToJSDateWithSameTimeAsSelectedTimeZone } from "@rilldata/web-common/lib/time/timezone";
import { timeFormat } from "d3-time-format";

export type SortFieldType = "measure" | "dimension" | "time";

export interface SortChip {
  label: string;
  type: SortFieldType;
}

export function getSortChips(
  sort: PivotSortTarget,
  config: PivotDataStoreConfig,
): SortChip[] {
  if ("field" in sort) {
    return [
      {
        label: getFieldLabel(sort.field, config),
        type: getSortFieldType(sort.field, config),
      },
    ];
  }

  return [
    ...sort.column_values.map(({ dimension, value }) => {
      const isTime = isTimeDimension(dimension, config.time.timeDimension);
      return {
        label: isTime ? formatColumnTimeValue(value, dimension, config) : value,
        type: isTime ? ("time" as const) : ("dimension" as const),
      };
    }),
    {
      label: getFieldLabel(sort.measure, config),
      type: "measure",
    },
  ];
}

function getSortFieldType(
  field: string,
  config: PivotDataStoreConfig,
): SortFieldType {
  if (isTimeDimension(field, config.time.timeDimension)) return "time";
  if (config.measureNames.includes(field)) return "measure";
  return "dimension";
}

const COMPARISON_MODIFIER: Record<string, string> = {
  [COMPARISON_VALUE]: " (previous)",
  [COMPARISON_DELTA]: " Δ",
  [COMPARISON_PERCENT]: " Δ %",
};

function formatColumnTimeValue(
  value: string,
  dimensionName: string,
  config: PivotDataStoreConfig,
): string {
  const grain = getTimeGrainFromDimension(dimensionName);
  const dt = convertISOStringToJSDateWithSameTimeAsSelectedTimeZone(
    value,
    config.time.timeZone || "UTC",
  );
  const formatter = timeFormat(grain ? TIME_GRAIN[grain].d3format : "%H:%M");
  return formatter(dt);
}

function getFieldLabel(field: string, config: PivotDataStoreConfig): string {
  if (isTimeDimension(field, config.time.timeDimension)) {
    const grain = getTimeGrainFromDimension(field);
    return `Time ${TIME_GRAIN[grain]?.label ?? grain}`;
  }

  for (const [suffix, modifier] of Object.entries(COMPARISON_MODIFIER)) {
    if (field.endsWith(suffix)) {
      const baseName = field.slice(0, -suffix.length);
      return getFieldLabel(baseName, config) + modifier;
    }
  }

  const measure = config.allMeasures.find((m) => m.name === field);
  if (measure) return measure.displayName || (measure.name as string);

  const dimension = config.allDimensions.find(
    (d) => (d.name || d.column) === field,
  );
  if (dimension) {
    return (
      dimension.displayName || dimension.name || (dimension.column as string)
    );
  }

  return field;
}
