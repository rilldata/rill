import type { DataModelerStateService } from "$common/data-modeler-state-service/DataModelerStateService";
import type {
    EntityStatus,
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import { EntityType, StateType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type { SinonSandbox, SinonStub } from "sinon";
import type { ApplicationStatus } from "$common/data-modeler-state-service/entity-state-service/ApplicationEntityService";

export class EntityStatusTracker {
    private statusChangeOrder: Array<EntityStatus>;
    private applicationStatusChangeOrder: Array<ApplicationStatus>;
    private stateSpyTimer: NodeJS.Timer;
    private originalStateService;
    private stateServiceSpy: SinonStub;

    public constructor(private readonly dataModelerStateService: DataModelerStateService,
                       private readonly sandbox: SinonSandbox) {
        this.originalStateService = dataModelerStateService.dispatch;
        this.stateServiceSpy = sandbox.stub(dataModelerStateService, "dispatch");
    }

    public init() {
        this.stateServiceSpy.callThrough();
    }

    public startTracker(entityType: EntityType) {
        this.statusChangeOrder = [];
        this.applicationStatusChangeOrder = [];
        // add some artificial delay to make sure we can catch status changes
        this.stateServiceSpy.callsFake(async (...args) => {
            return this.originalStateService.apply(this.dataModelerStateService, args);
        });
        this.stateSpyTimer = setInterval(() => {
            const entity = this.dataModelerStateService
                .getEntityStateService(entityType, StateType.Derived)
                .getCurrentState().entities[0];
            if (entity && this.statusChangeOrder[this.statusChangeOrder.length - 1] !== entity.status) {
                this.statusChangeOrder.push(entity.status);
            }

            const applicationState = this.dataModelerStateService
                .getEntityStateService(EntityType.Application, StateType.Derived)
                .getCurrentState();
            if (this.applicationStatusChangeOrder[this.applicationStatusChangeOrder.length - 1] !== applicationState.status) {
                this.applicationStatusChangeOrder.push(applicationState.status);
            }
        });
    }

    public stopTracker() {
        if (this.stateSpyTimer) clearInterval(this.stateSpyTimer);
    }

    public getStatusChangeOrder() {
        return this.statusChangeOrder;
    }
    public getApplicationStatusChangeOrder() {
        return this.applicationStatusChangeOrder;
    }
}
