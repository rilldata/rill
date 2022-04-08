import { QueryTreeNode } from "./QueryTreeNode";
import type { ColumnNode } from "./ColumnNode";
import type { TableNode } from "./TableNode";

export class SelectNode extends QueryTreeNode {
    public tables = new Array<TableNode>();
    public columns = new Array<ColumnNode>();
}
