import type { CTENode } from "./CTENode";
import type { QueryTreeNodeJSON } from "./QueryTreeNode";
import type { SelectNode } from "./SelectNode";
import type { TableNode, TableNodeJSON } from "./TableNode";

export interface QueryTreeJSON {
    root: QueryTreeNodeJSON;
    tables: Array<TableNodeJSON>;
}

export class QueryTree {
    public root: SelectNode | CTENode;
    public tables = new Array<TableNode>();

    public addTable(tableNode: TableNode) {
        this.tables.push(tableNode);
    }

    public toJSON(): QueryTreeJSON {
        return {
            root: this.root.toJSON(),
            tables: this.tables.map(table => table.toJSON()),
        };
    }
}
