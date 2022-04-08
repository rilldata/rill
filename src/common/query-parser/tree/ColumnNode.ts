import type { ColumnRefNode } from "./ColumnRefNode";
import { QueryTreeNode } from "./QueryTreeNode";
import type { TableNode } from "./TableNode";

export class ColumnNode extends QueryTreeNode {
    public table: TableNode;
    public columnRefs = new Array<ColumnRefNode>();
    public alias: string;
}
