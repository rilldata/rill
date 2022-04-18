import type { QueryTreeNode } from "./QueryTreeNode";

export class NodeStack<TreeNode extends QueryTreeNode> {
    public stack = new Array<TreeNode>();
    public currentNode: TreeNode;

    public enterNode(node: TreeNode) {
        if (this.currentNode) this.stack.push(this.currentNode);
        this.currentNode = node as TreeNode;
    }

    public exitNode(node: TreeNode) {
        if (this.currentNode !== node) return;
        this.currentNode = this.stack.pop();
    }
}
