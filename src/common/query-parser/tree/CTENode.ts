import { QueryTreeNode, QueryTreeNodeJSON } from "./QueryTreeNode";
import { QueryTreeNodeType } from "./QueryTreeNodeType";
import type { SelectNode, SelectNodeJSON } from "./SelectNode";
import type { TableNode, TableNodeJSON } from "./TableNode";

export interface CTENodeJSON extends QueryTreeNodeJSON {
    tables: Array<TableNodeJSON>;
    select: SelectNodeJSON;
}

export class CTENode extends QueryTreeNode {
    public readonly type = QueryTreeNodeType.CTE;
    public tables = new Array<TableNode>();
    public select: SelectNode;

    public toJSON(includeLocation = false) {
        return {
            ...super.toJSON(includeLocation),
            tables: this.tables.map(table => table.toJSON(includeLocation)),
            ...this.select ? {select: this.select.toJSON(includeLocation)} : {},
        }
    }
}
