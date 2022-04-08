import type { CTENode } from "./CTENode";
import type { SelectNode } from "./SelectNode";
import type { TableNode } from "./TableNode";

export class QueryTree {
    public root: SelectNode | CTENode;
    public tables = new Array<TableNode>();

    public addTable(tableNode: TableNode) {
        this.tables.push(tableNode);
    }
}
