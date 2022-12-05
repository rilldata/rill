import { Heap } from "@rilldata/web-local/lib/http-request-queue/Heap";

export interface ActionMetadata {
  priority: number;
  id: string;
}

export interface ActionPromiseCallbacks {
  promiseResolve: (arg: unknown) => void;
  promiseReject: (error: Error) => void;
}
export type QueuedAction = [
  action: string,
  args: Array<unknown>,
  callbacks: ActionPromiseCallbacks
];
export const QueuedActionNameIdx = 0;
export const QueuedActionArgsIdx = 1;
export const QueuedActionCallbacksIdx = 2;

export interface QueueEntry {
  metadata: ActionMetadata;
  actions: Array<QueuedAction>;
}

const heapCompareFunc = (a: QueueEntry, b: QueueEntry) => {
  return b.metadata.priority - a.metadata.priority;
};

export class PriorityActionQueue {
  private maxHeap = new Heap<QueueEntry>(heapCompareFunc);
  private heapEntryMap = new Map<string, QueueEntry>();

  public enqueue(
    actionMetadata: ActionMetadata,
    queuedAction: QueuedAction
  ): void {
    if (this.heapEntryMap.has(actionMetadata.id)) {
      const existingItem: QueueEntry = this.heapEntryMap.get(actionMetadata.id);
      existingItem.actions.push(queuedAction);
      this.maxHeap.updateItem(existingItem);
    } else {
      const newItem: QueueEntry = {
        metadata: actionMetadata,
        actions: [queuedAction],
      };
      this.maxHeap.push(newItem);
      this.heapEntryMap.set(actionMetadata.id, newItem);
    }
  }

  public clearQueue(id: string): Array<QueuedAction> {
    if (!this.heapEntryMap.has(id)) return;
    const existingItem: QueueEntry = this.heapEntryMap.get(id);
    const actions = existingItem.actions;
    // clear the actions in an entry.
    // dequeue will clear this from heap
    existingItem.actions = [];
    return actions;
  }

  public updatePriority(id: string, priority: number): void {
    if (!this.heapEntryMap.has(id)) return;
    const existingItem: QueueEntry = this.heapEntryMap.get(id);
    existingItem.metadata.priority = priority;
    this.maxHeap.updateItem(existingItem);
  }

  public dequeue(): QueuedAction {
    if (this.maxHeap.empty()) return undefined;

    let topItem: QueueEntry = this.maxHeap.peek();
    // remove any entry that has empty actions
    while (topItem.actions.length === 0) {
      this.maxHeap.pop();
      this.heapEntryMap.delete(topItem.metadata.id);
      if (this.maxHeap.empty()) return undefined;
      topItem = this.maxHeap.peek();
    }
    if (topItem.actions.length === 1) {
      this.maxHeap.pop();
      this.heapEntryMap.delete(topItem.metadata.id);
    }
    return topItem.actions.shift();
  }
}
