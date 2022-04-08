import { QueryTreeNode } from "./QueryTreeNode";
import type { ColumnNode } from "./ColumnNode";

export class TableNode extends QueryTreeNode {
    public tableName: string;
    public alias: string;
    public availableColumns = new Array<ColumnNode>();
}
