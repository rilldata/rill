import type { CTENode } from "./CTENode";
import type { QueryTreeNodeJSON } from "./QueryTreeNode";
import type { SelectNode } from "./SelectNode";
import type { TableNode, TableNodeJSON } from "./TableNode";

export interface QueryTreeJSON {
  root: QueryTreeNodeJSON;
  sourceTables: Array<TableNodeJSON>;
}

export class QueryTree {
  public root: SelectNode | CTENode;
  public sourceTables = new Array<TableNode>();

  public addTable(tableNode: TableNode) {
    this.sourceTables.push(tableNode);
  }

  public toJSON(): QueryTreeJSON {
    return {
      root: this.root.toJSON(),
      sourceTables: this.sourceTables.map((table) => table.toJSON()),
    };
  }
}
