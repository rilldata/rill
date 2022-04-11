import { QueryTreeNode, QueryTreeNodeJSON } from "./QueryTreeNode";
import type { ColumnNode } from "./ColumnNode";
import { QueryTreeNodeType } from "./QueryTreeNodeType";

export interface TableNodeJSON extends QueryTreeNodeJSON {
    tableName?: string;
    alias?: string;
    isSourceTable: boolean;
}

export class TableNode extends QueryTreeNode {
    public readonly type: QueryTreeNodeType = QueryTreeNodeType.Table;
    public tableName: string;
    public alias: string;
    public availableColumns = new Array<ColumnNode>();
    public isSourceTable = false;

    public toJSON(includeLocation = false): TableNodeJSON {
        return {
            ...super.toJSON(includeLocation),
            ...this.tableName ? {tableName: this.tableName} : {},
            ...this.alias ? {alias: this.alias} : {},
            isSourceTable: this.isSourceTable,
        };
    }
}
