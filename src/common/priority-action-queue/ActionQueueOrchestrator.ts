import type { ActionServiceBase } from "$common/ServiceBase";
import {
    PriorityActionQueue, QueuedActionArgsIdx,
    QueuedActionCallbacksIdx, QueuedActionNameIdx
} from "$common/priority-action-queue/PriorityActionQueue";
import type {
    ActionMetadata
} from "$common/priority-action-queue/PriorityActionQueue";

export class ActionQueueOrchestrator<ActionsDefinition extends Record<string, Array<any>>> {
    private actionService: ActionServiceBase<ActionsDefinition>;
    private priorityActionQueue = new PriorityActionQueue();

    private running = false;

    public constructor(actionService: ActionServiceBase<ActionsDefinition>) {
        this.actionService = actionService;
    }

    public async run(): Promise<void> {
        if (this.running) return;
        this.running = true;

        let queuedAction = this.priorityActionQueue.dequeue();
        while (queuedAction !== undefined) {
            try {
                const result = await this.actionService.dispatch(
                    queuedAction[QueuedActionNameIdx], queuedAction[QueuedActionArgsIdx] as any);
                queuedAction[QueuedActionCallbacksIdx].promiseResolve(result);
            } catch (err) {
                queuedAction[QueuedActionCallbacksIdx].promiseReject(err);
            }
            queuedAction = this.priorityActionQueue.dequeue();
        }

        this.running = false;
    }

    public async enqueue<Action extends keyof ActionsDefinition>(
        actionMetadata: ActionMetadata, action: Action, args: ActionsDefinition[Action],
    ): Promise<any> {
        return new Promise((resolve, reject) => {
            this.priorityActionQueue.enqueue(actionMetadata, [action as any, args, {
                promiseResolve: resolve, promiseReject: reject,
            }]);
            if (!this.running) {
                setTimeout(() => this.run());
            }
        });
    }

    public clearQueue(id: string): void {
        const queuedActions = this.priorityActionQueue.clearQueue(id);
        if (!queuedActions) return;
        queuedActions.forEach(queuedAction =>
            queuedAction[QueuedActionCallbacksIdx].promiseReject(new Error("Cancelled")));
    }

    public updatePriority(id: string, priority: number): void {
        this.priorityActionQueue.updatePriority(id, priority);
    }
}
