import PercentageChange from "@rilldata/web-common/components/data-types/PercentageChange.svelte";
import PivotExpandableCell from "@rilldata/web-common/features/dashboards/pivot//PivotExpandableCell.svelte";
import PivotMeasureCell from "@rilldata/web-common/features/dashboards/pivot//PivotMeasureCell.svelte";
import {
  formatRowDimensionValue,
  getDimensionColumnProps,
  getMeasureColumnProps,
} from "@rilldata/web-common/features/dashboards/pivot/pivot-column-definition";
import { cellComponent } from "@rilldata/web-common/features/dashboards/pivot/pivot-utils";
import {
  PivotDataRow,
  PivotDataStoreConfig,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import { formatMeasurePercentageDifference } from "@rilldata/web-common/lib/number-formatting/percentage-formatter";
import { ColumnDef } from "@tanstack/svelte-table";

export function getColumnDefForTDD(config: PivotDataStoreConfig) {
  const { rowDimensionNames } = config;

  const measures = getMeasureColumnProps(config);
  const rowDimensions = getDimensionColumnProps(rowDimensionNames, config);

  let rowDimensionsForColumnDef = rowDimensions;

  rowDimensionsForColumnDef = rowDimensions.slice(0, 1);
  const nestedLabel = rowDimensions.map((d) => d.label || d.name).join(" > ");

  const rowDefinitions: ColumnDef<PivotDataRow>[] =
    rowDimensionsForColumnDef.map((d) => {
      return {
        id: d.name,
        accessorFn: (row) => row[d.name],
        header: nestedLabel,
        cell: ({ row, getValue }) =>
          cellComponent(PivotExpandableCell, {
            value: formatRowDimensionValue(
              getValue() as string,
              row.depth,
              config.time,
              rowDimensionNames,
            ),
            row,
          }),
      };
    });

  const leafColumns: (ColumnDef<PivotDataRow> & { name: string })[] =
    measures.map((m) => {
      return {
        accessorKey: m.name,
        header: m.label || m.name,
        name: m.name,
        cell: (info) => {
          const measureValue = info.getValue() as number | null | undefined;
          if (m.type === "comparison_percent") {
            return cellComponent(PercentageChange, {
              isNull: measureValue == null,
              value:
                measureValue !== null && measureValue !== undefined
                  ? formatMeasurePercentageDifference(measureValue)
                  : null,
              inTable: true,
            });
          }
          const value = m.formatter(measureValue);

          if (value == null) return cellComponent(PivotMeasureCell, {});
          return value;
        },
      };
    });

  return [...rowDefinitions, ...leafColumns];
}
