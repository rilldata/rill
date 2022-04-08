import {astVisitor, ExprRef, FromStatement, FromTable, IAstVisitor, parse, SelectedColumn, SelectFromStatement, WithRecursiveStatement, WithStatement} from "pgsql-ast-parser";
import { QueryTreeTracker } from "./QueryTreeTarcker";
import { QueryTree } from "./tree/QueryTree";

export class QueryParser {
    private readonly queryTree = new QueryTree();
    private readonly queryTreeTracker = new QueryTreeTracker(this.queryTree);

    public parse(query: string): QueryTree {
        const visitor = astVisitor(map => ({
            selection: sel => {
                this.handleSelection(map, sel);
            },

            with: w => {
                this.handleCTE(map, w);
            },
            withRecursive: wr => {
                this.handleRecursiveCTE(map, wr);
            },

            fromTable: ft => {
                this.handleFromTable(map, ft);
            },
            fromStatement: fs => {
                this.handleSubQuery(map, fs);
            },
            fromCall: fc => {
                map.super().fromCall(fc);
                // TODO
            },

            selectionColumn: col => {
                this.handleSelectionColumn(map, col);
            },
            
            ref: r => {
                this.handleRef(map, r);
            },
        }));
        visitor.statement(parse(query, { locationTracking: true })[0]);

        return this.queryTree;
    }

    private handleSelection(visitor: IAstVisitor, selection: SelectFromStatement) {
        this.queryTreeTracker.enterSelection(selection);
        visitor.super().selection(selection);
        this.queryTreeTracker.exitNode();
    }

    private handleFromTable(visitor: IAstVisitor, fromTable: FromTable) {
        this.queryTreeTracker.enterTable(fromTable);
        visitor.super().fromTable(fromTable);
        this.queryTreeTracker.exitNode();
    }
    private handleSubQuery(visitor: IAstVisitor, fromStatement: FromStatement) {
        this.queryTreeTracker.enterSubQuery(fromStatement);
        visitor.super().fromStatement(fromStatement);
        this.queryTreeTracker.exitNode();
    }

    private handleCTE(visitor: IAstVisitor, withStatement: WithStatement) {
        this.queryTreeTracker.enterCTE(withStatement);
        withStatement.bind.forEach((bind) => {
            this.queryTreeTracker.enterCTETable(bind.alias.name, bind.statement);
            visitor.super().statement(bind.statement);
            this.queryTreeTracker.exitNode();
        });
        visitor.super().statement(withStatement.in);
        this.queryTreeTracker.exitNode();
    }
    private handleRecursiveCTE(visitor: IAstVisitor, withStatement: WithRecursiveStatement) {
        this.queryTreeTracker.enterCTE(withStatement);
        // TODO
        visitor.super().statement(withStatement.in);
        this.queryTreeTracker.exitNode();
    }

    private handleSelectionColumn(visitor: IAstVisitor, selectionColumn: SelectedColumn) {
        this.queryTreeTracker.enterColumn(selectionColumn);
        visitor.super().selectionColumn(selectionColumn);
        this.queryTreeTracker.exitNode();
    }

    private handleRef(visitor: IAstVisitor, ref: ExprRef) {
        this.queryTreeTracker.handleRef(ref);
        visitor.super().ref(ref);
    }
}
