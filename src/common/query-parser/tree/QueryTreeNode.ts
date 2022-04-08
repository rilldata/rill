import type { NodeLocation } from "pgsql-ast-parser";

export class QueryTreeNode {
    public readonly start: number;
    public readonly end: number;

    public constructor(location: NodeLocation) {
        this.start = location?.start ?? -1;
        this.end = location?.end ?? -1;
    }
}
