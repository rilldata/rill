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
        this.start = location?.start ?? undefined;
        this.end = location?.end ?? undefined;
    }

    public toJSON(): QueryTreeNodeJSON {
        return {
            type: this.type,
            start: this.start,
            end: this.end,
        };
    }
}
