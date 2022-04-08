import { QueryTreeNode } from "./QueryTreeNode";
import type { SelectNode } from "./SelectNode";
import type { TableNode } from "./TableNode";

export class CTENode extends QueryTreeNode {
    public tables = new Array<TableNode>();
    public select: SelectNode;
}
