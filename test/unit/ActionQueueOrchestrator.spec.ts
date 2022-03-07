import { TestBase } from "@adityahegde/typescript-test-utils";
import { JestTestLibrary } from "@adityahegde/typescript-test-utils/dist/jest/JestTestLibrary";
import type { SinonStub } from "sinon";
import { ActionQueueOrchestrator } from "$common/priority-action-queue/ActionQueueOrchestrator";
import { asyncWait } from "$common/utils/waitUtils";

interface TestActionsDefinition extends Record<string, any> {
    "actionOne": [string, string],
    "actionTwo": [string, string],
    "actionThree": [string, string],
}
enum TestPriorities {
    One,
    Two,
    Three,
}

const ID1 = "1";
const ID2 = "2";
const ID3 = "3";
const EXPECTED_RUN_ORDER = [
    [ "actionOne", [ "1", "one" ] ],
    [ "actionOne", [ "2", "one" ] ],
    [ "actionTwo", [ "2", "two" ] ],
    [ "actionThree", [ "2", "three" ] ],
    [ "actionTwo", [ "1", "two" ] ],
    [ "actionThree", [ "1", "three" ] ],
];

@TestBase.Suite
@TestBase.TestLibrary(JestTestLibrary)
export class ActionQueueOrchestratorSpec extends TestBase {
    private actionService: { dispatch: SinonStub };
    private actionQueueOrchestrator: ActionQueueOrchestrator<TestActionsDefinition>;

    @TestBase.BeforeSuite()
    public setup() {
        this.actionService = { dispatch: this.sandbox.stub() };
        this.actionQueueOrchestrator =
            new ActionQueueOrchestrator(this.actionService);
    }

    @TestBase.BeforeEachTest()
    public setupTests() {
        this.actionService.dispatch.callsFake(() => asyncWait(100));
    }

    @TestBase.Test()
    public async shouldRunHigherPriorityActions() {
        const promisesOne = this.getActionsForID(ID1, TestPriorities.Three);
        await asyncWait(50);
        const promisesTwo = this.getActionsForID(ID2, TestPriorities.One);
        await Promise.all([
            ...promisesOne,
            ...promisesTwo,
        ]);

        // after the 1st action the queue switches to ID2
        expect(this.actionService.dispatch.args).toStrictEqual(EXPECTED_RUN_ORDER);
    }

    @TestBase.Test()
    public async shouldRunAlteredPriorityID() {
        const promisesOne = this.getActionsForID(ID1, TestPriorities.Two);
        const promisesTwo = this.getActionsForID(ID2, TestPriorities.Three);
        await asyncWait(50);
        this.actionQueueOrchestrator.updatePriority(ID2, TestPriorities.One);
        await Promise.all([
            ...promisesOne,
            ...promisesTwo,
        ]);

        // after the 1st action the queue switches to ID2 as its priority increased
        expect(this.actionService.dispatch.args).toStrictEqual(EXPECTED_RUN_ORDER);
    }

    @TestBase.Test()
    public async shouldSwitchFromRunningAlteredPriorityID() {
        const promisesOne = this.getActionsForID(ID1, TestPriorities.One);
        const promisesTwo = this.getActionsForID(ID2, TestPriorities.Two);
        await asyncWait(50);
        this.actionQueueOrchestrator.updatePriority(ID1, TestPriorities.Three);
        await Promise.all([
            ...promisesOne,
            ...promisesTwo,
        ]);

        // after the 1st action the queue switches to ID2 as ID1"s priority lowered
        expect(this.actionService.dispatch.args).toStrictEqual(EXPECTED_RUN_ORDER);
    }

    @TestBase.Test()
    public async shouldStopRunningCancelledQueries() {
        const promisesOne = this.getActionsForID(ID1, TestPriorities.One);
        const promisesTwo = this.getActionsForID(ID2, TestPriorities.Three);
        await asyncWait(50);
        this.actionQueueOrchestrator.clearQueue(ID1);
        await asyncWait(75);
        const promisesThree = this.getActionsForID(ID3, TestPriorities.Two);
        await Promise.all([
            ...promisesOne,
            ...promisesTwo,
            ...promisesThree
        ]);

        // after the 1st action ID1 is cancelled, ID2 picks up
        // after 1st action of ID2, ID3 is picked as it has higher priority
        expect(this.actionService.dispatch.args).toStrictEqual([
            [ "actionOne", [ "1", "one" ] ],
            [ "actionOne", [ "2", "one" ] ],
            [ "actionOne", [ "3", "one" ] ],
            [ "actionTwo", [ "3", "two" ] ],
            [ "actionThree", [ "3", "three" ] ],
            [ "actionTwo", [ "2", "two" ] ],
            [ "actionThree", [ "2", "three" ] ]
        ]);
    }

    private getActionsForID(id: string, priority: TestPriorities) {
        return [
            this.wrapEnqueue(id, priority, "actionOne", [id, "one"]),
            this.wrapEnqueue(id, priority, "actionTwo", [id, "two"]),
            this.wrapEnqueue(id, priority, "actionThree", [id, "three"]),
        ];
    }
    private async wrapEnqueue<Action extends keyof TestActionsDefinition>(
        id: string, priority: TestPriorities,
        action: keyof TestActionsDefinition, args: TestActionsDefinition[Action],
    ) {
        try {
            await this.actionQueueOrchestrator.enqueue({id, priority}, action, args);
            // eslint-disable-next-line no-empty
        } catch (err) {}
    }
}
