import { QueryTreeNode, QueryTreeNodeJSON } from "./QueryTreeNode";
import type { ColumnNode, ColumnNodeJSON } from "./ColumnNode";
import type { TableNode, TableNodeJSON } from "./TableNode";
import { QueryTreeNodeType } from "./QueryTreeNodeType";

export interface SelectNodeJSON extends QueryTreeNodeJSON {
    tables: Array<TableNodeJSON>;
    columns: Array<ColumnNodeJSON>;
}

export class SelectNode extends QueryTreeNode {
    public readonly type = QueryTreeNodeType.Select;
    public tables = new Array<TableNode>();
    public columns = new Array<ColumnNode>();

    public toJSON(includeLocation = false): SelectNodeJSON {
        return {
            ...super.toJSON(includeLocation),
            tables: this.tables.map(table => table.toJSON()),
            columns: this.columns.map(column => column.toJSON()),
        };
    }
}
