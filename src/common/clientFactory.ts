import {DataModelerStateService} from "$common/data-modeler-state-service/DataModelerStateService";
import type {DataModelerService} from "$common/data-modeler-service/DataModelerService";
import {DataModelerSocketService} from "$common/socket/DataModelerSocketService";
import type {RootConfig} from "$common/config/RootConfig";
import {
    PersistentTableEntityService
} from "$common/data-modeler-state-service/entity-state-service/PersistentTableEntityService";
import {
    DerivedTableEntityService
} from "$common/data-modeler-state-service/entity-state-service/DerivedTableEntityService";
import {
    PersistentModelEntityService
} from "$common/data-modeler-state-service/entity-state-service/PersistentModelEntityService";
import {
    DerivedModelEntityService
} from "$common/data-modeler-state-service/entity-state-service/DerivedModelEntityService";
import {
    ApplicationStateService
} from "$common/data-modeler-state-service/entity-state-service/ApplicationEntityService";

export function dataModelerStateServiceClientFactory() {
    return new DataModelerStateService([],
        [
            PersistentTableEntityService, DerivedTableEntityService,
            PersistentModelEntityService, DerivedModelEntityService,
            ApplicationStateService,
        ].map(EntityStateService => new EntityStateService()));
}

export function clientFactory(config: RootConfig): {
    dataModelerStateService: DataModelerStateService,
    dataModelerService: DataModelerService,
} {
    const dataModelerStateService = dataModelerStateServiceClientFactory();
    const dataModelerService = new DataModelerSocketService(dataModelerStateService, config.server);

    return {dataModelerStateService, dataModelerService};
}
