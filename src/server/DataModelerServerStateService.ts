import { DataModelerStateService } from "$common/data-modeler-state-service/DataModelerStateService";
import type { DataModelerState } from "$lib/types";
import { writable } from "svelte/store";
import { initialState } from "$common/stateInstancesFactory";
import { existsSync, readFileSync, writeFileSync } from "fs";

/**
 * State service class present on the server.
 * Reads the initial state from file and syncs it back.
 */
export class DataModelerServerStateService extends DataModelerStateService {
    private timer: NodeJS.Timer;

    public init(): void {
        let initState: DataModelerState;

        if (this.config.state.autoSync && existsSync(this.config.state.savedStateFile)) {
            initState = JSON.parse(
                readFileSync(this.config.state.savedStateFile).toString());
        } else {
            initState = initialState();
        }
        this.store = writable(initState);

        if (this.config.state.autoSync) {
            this.periodicallySyncStateToFile();
        }
    }

    public destroy(): void {
        this.syncStateToFile();
        if (this.timer) {
            clearInterval(this.timer);
        }
    }

    private periodicallySyncStateToFile() {
        this.timer = setInterval(() => {
            this.syncStateToFile();
        }, 500);
    }

    private syncStateToFile() {
        writeFileSync(this.config.state.savedStateFile,
            JSON.stringify(this.getCurrentState()));
    }
}
