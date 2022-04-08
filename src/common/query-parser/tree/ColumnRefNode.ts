import { QueryTreeNode } from "./QueryTreeNode";
import type { TableNode } from "./TableNode";

export class ColumnRefNode extends QueryTreeNode {
    public table: TableNode;
    public name: string;
    public fullName: string;
}
