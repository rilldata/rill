import type { ColumnNodeJSON } from "$common/query-parser/tree/ColumnNode";
import type { ColumnRefNodeJSON } from "$common/query-parser/tree/ColumnRefNode";
import type { CTENodeJSON } from "$common/query-parser/tree/CTENode";
import type { NestedSelectNodeJSON } from "$common/query-parser/tree/NestedSelectNode";
import type { QueryTreeJSON } from "$common/query-parser/tree/QueryTree";
import type { QueryTreeNodeJSON } from "$common/query-parser/tree/QueryTreeNode";
import { QueryTreeNodeType } from "$common/query-parser/tree/QueryTreeNodeType";
import type { SelectNodeJSON } from "$common/query-parser/tree/SelectNode";
import type { SourceNodeJSON } from "$common/query-parser/tree/SourceNode";

export function expectedQueryTree(
  root: QueryTreeNodeJSON,
  sources: Array<SourceNodeJSON>
): QueryTreeJSON {
  return {
    root,
    sources,
  };
}

export function expectedCTE(
  sources: Array<SourceNodeJSON>,
  select: SelectNodeJSON
): CTENodeJSON {
  return {
    type: QueryTreeNodeType.CTE,
    sources,
    select,
  };
}

export function expectedSelect(
  sources: Array<SourceNodeJSON>,
  columns: Array<ColumnNodeJSON>
): SelectNodeJSON {
  return {
    type: QueryTreeNodeType.Select,
    sources,
    columns,
  };
}

export function expectedSource(
  sourceName: string,
  alias?: string
): SourceNodeJSON {
  return {
    type: QueryTreeNodeType.Source,
    sourceName,
    alias: alias ?? sourceName,
    isSource: false,
  };
}
export function expectedSource(
  sourceName: string,
  alias?: string
): SourceNodeJSON {
  return {
    ...expectedSource(sourceName, alias),
    isSource: true,
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
    isSource: false,
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
