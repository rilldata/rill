import type { NodeLocation } from "pgsql-ast-parser";
import type { QueryTreeNodeType } from "./QueryTreeNodeType";

export interface QueryTreeNodeJSON {
    type: QueryTreeNodeType;
    start?: number;
    end?: number;
}

export class QueryTreeNode {
    public readonly type: QueryTreeNodeType;
    public readonly start: number;
    public readonly end: number;

    public constructor(location: NodeLocation) {
        this.start = location?.start ?? -1;
        this.end = location?.end ?? -1;
    }

    public toJSON(includeLocation = false): QueryTreeNodeJSON {
        return {
            type: this.type,
            ...includeLocation ? {
                start: this.start,
                end: this.end,
            } : {},
        };
    }
}
