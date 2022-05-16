import type { ColumnRefNode, ColumnRefNodeJSON } from "./ColumnRefNode";
import { QueryTreeNode, QueryTreeNodeJSON } from "./QueryTreeNode";
import { QueryTreeNodeType } from "./QueryTreeNodeType";

export interface ColumnNodeJSON extends QueryTreeNodeJSON {
  refs: Array<ColumnRefNodeJSON>;
  alias?: string;
}

export class ColumnNode extends QueryTreeNode {
  public readonly type = QueryTreeNodeType.Column;
  public columnRefs = new Array<ColumnRefNode>();
  public alias: string;

  public toJSON(): ColumnNodeJSON {
    return {
      ...super.toJSON(),
      refs: this.columnRefs.map((columnRef) => columnRef.toJSON()),
      ...(this.alias ? { alias: this.alias } : {}),
    };
  }
}
