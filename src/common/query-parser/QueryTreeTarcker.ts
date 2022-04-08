import type { SelectFromStatement, FromTable, FromStatement, SelectedColumn, ExprRef } from "pgsql-ast-parser";
import { ColumnNode } from "./tree/ColumnNode";
import { ColumnRefNode } from "./tree/ColumnRefNode";
import { NodeStack } from "./tree/NodeStack";
import type { QueryTree } from "./tree/QueryTree";
import type { QueryTreeNode } from "./tree/QueryTreeNode";
import { SelectNode } from "./tree/SelectNode";
import { TableNode } from "./tree/TableNode";

export class QueryTreeTracker {
    private readonly fullStack = new NodeStack<QueryTreeNode>();
    private readonly selectStack = new NodeStack<SelectNode>();
    private readonly tableStack = new NodeStack<TableNode>();
    private readonly columnStack = new NodeStack<ColumnNode>();

    public tableMap = new Map<string, Array<TableNode>>();

    public constructor(private readonly queryTree: QueryTree) {}

    public exitNode() {
        this.selectStack.exitNode(this.fullStack.currentNode as SelectNode);
        this.tableStack.exitNode(this.fullStack.currentNode as TableNode);
        this.columnStack.exitNode(this.fullStack.currentNode as ColumnNode);

        this.fullStack.exitNode(this.fullStack.currentNode);
    }

    public enterSelection(select: SelectFromStatement) {
        const selectNode = new SelectNode(select._location);
        
        if (!this.queryTree.root) {
            this.queryTree.root = selectNode;
        }

        this.selectStack.enterNode(selectNode);
        this.fullStack.enterNode(selectNode);
    }

    public enterTable(table: FromTable) {
        const tableNode = new TableNode(table._location);
        tableNode.tableName = table.name.name;
        tableNode.alias = table.name.alias ?? table.name.name;

        this.handleTableNode(tableNode);
    }
    public enterSubQuery(fromStatement: FromStatement) {
        const tableNode = new TableNode(fromStatement._location);
        tableNode.alias = fromStatement.alias;

        this.handleTableNode(tableNode);
    }

    public exitTable() {
        if (!this.isAtTable()) return;

        const name = this.tableStack.currentNode.alias ??
            this.tableStack.currentNode.tableName;
        this.tableMap.get(name).pop();
    }

    public enterColumn(column: SelectedColumn) {
        const columnNode = new ColumnNode(column._location);
        if (column.alias) columnNode.alias = column.alias.name;

        if (this.isAtSelect()) {
            this.selectStack.currentNode.columns.push(columnNode);
        }

        this.columnStack.enterNode(columnNode);
        this.fullStack.enterNode(columnNode);
    }

    public handleRef(ref: ExprRef) {
        if (this.isAtColumn()) {
            this.handleColumnRef(ref);
        }
    }

    private handleTableNode(tableNode: TableNode) {
        const name = tableNode.alias ?? tableNode.tableName;
        if (!this.tableMap.has(name)) {
            this.tableMap.set(name, [tableNode]);
        } else {
            this.tableMap.get(name).push(tableNode);
        }

        if (this.isAtSelect()) {
            this.selectStack.currentNode.tables.push(tableNode);
        }

        this.tableStack.enterNode(tableNode);
        this.fullStack.enterNode(tableNode);
        this.queryTree.addTable(tableNode);
    }

    private handleColumnRef(ref: ExprRef) {
        const colRefNode = new ColumnRefNode(ref._location);
        colRefNode.name = ref.name;
        colRefNode.fullName = "";

        if (ref.table && this.tableMap.has(ref.table.name)) {
            colRefNode.table = this.tableMap.get(ref.table.name)[
                this.tableMap.get(ref.table.name).length - 1
            ];
            colRefNode.fullName = `${colRefNode.table.alias}.`;
        }

        colRefNode.fullName += colRefNode.name;
        this.columnStack.currentNode.columnRefs.push(colRefNode);
    }

    private isAtColumn() {
        return this.columnStack.currentNode === this.fullStack.currentNode;
    }
    private isAtTable() {
        return this.tableStack.currentNode === this.fullStack.currentNode;
    }
    private isAtSelect() {
        return this.selectStack.currentNode === this.fullStack.currentNode;
    }
}
