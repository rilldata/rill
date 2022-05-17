import type { ColumnNodeJSON } from "$common/query-parser/tree/ColumnNode";
import type { ColumnRefNodeJSON } from "$common/query-parser/tree/ColumnRefNode";
import type { CTENodeJSON } from "$common/query-parser/tree/CTENode";
import type { NestedSelectNodeJSON } from "$common/query-parser/tree/NestedSelectNode";
import type { QueryTreeJSON } from "$common/query-parser/tree/QueryTree";
import type { QueryTreeNodeJSON } from "$common/query-parser/tree/QueryTreeNode";
import { QueryTreeNodeType } from "$common/query-parser/tree/QueryTreeNodeType";
import type { SelectNodeJSON } from "$common/query-parser/tree/SelectNode";
import type { TableNodeJSON } from "$common/query-parser/tree/TableNode";

export function expectedQueryTree(
  root: QueryTreeNodeJSON,
  sourceTables: Array<TableNodeJSON>
): QueryTreeJSON {
  return {
    root,
    sourceTables,
  };
}

export function expectedCTE(
  tables: Array<TableNodeJSON>,
  select: SelectNodeJSON
): CTENodeJSON {
  return {
    type: QueryTreeNodeType.CTE,
    tables,
    select,
  };
}

export function expectedSelect(
  tables: Array<TableNodeJSON>,
  columns: Array<ColumnNodeJSON>
): SelectNodeJSON {
  return {
    type: QueryTreeNodeType.Select,
    tables,
    columns,
  };
}

export function expectedTable(
  tableName: string,
  alias?: string
): TableNodeJSON {
  return {
    type: QueryTreeNodeType.Table,
    tableName,
    alias: alias ?? tableName,
    isSourceTable: false,
  };
}
export function expectedSourceTable(
  tableName: string,
  alias?: string
): TableNodeJSON {
  return {
    ...expectedTable(tableName, alias),
    isSourceTable: true,
  };
}
export function expectedNestedSelect(
  select: SelectNodeJSON,
  alias: string
): NestedSelectNodeJSON {
  return {
    type: QueryTreeNodeType.NestedSelect,
    select,
    alias,
    isSourceTable: false,
  };
}

export function expectedColumnRef(fullName: string): ColumnRefNodeJSON {
  return {
    type: QueryTreeNodeType.ColumnRef,
    fullName,
  };
}
export function expectedColumn(
  refs: Array<string>,
  alias?: string
): ColumnNodeJSON {
  return {
    type: QueryTreeNodeType.Column,
    refs: refs.map(expectedColumnRef),
    ...(alias ? { alias } : {}),
  };
}
