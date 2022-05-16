import { QueryTreeNode, QueryTreeNodeJSON } from "./QueryTreeNode";
import { QueryTreeNodeType } from "./QueryTreeNodeType";
import type { TableNode } from "./TableNode";

export interface ColumnRefNodeJSON extends QueryTreeNodeJSON {
  fullName: string;
}

export class ColumnRefNode extends QueryTreeNode {
  public readonly type = QueryTreeNodeType.ColumnRef;
  public table: TableNode;
  public name: string;
  public fullName: string;

  public toJSON(): ColumnRefNodeJSON {
    return {
      ...super.toJSON(),
      fullName: this.fullName,
    };
  }
}
