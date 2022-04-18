import { QueryTreeNodeType } from "./QueryTreeNodeType";
import type { SelectNode, SelectNodeJSON } from "./SelectNode";
import { TableNode, TableNodeJSON } from "./TableNode";

export interface NestedSelectNodeJSON extends TableNodeJSON {
    select: SelectNodeJSON;
}

export class NestedSelectNode extends TableNode {
    public readonly type = QueryTreeNodeType.NestedSelect;
    public select: SelectNode;

    public toJSON(): NestedSelectNodeJSON {
        return {
            ...super.toJSON(),
            select: this.select.toJSON(),
        }
    }
}
