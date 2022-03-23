import axios from "axios";
import type { RootConfig } from "$common/config/RootConfig";
import type { MetricsEvent } from "$common/metrics/MetricsTypes";

export class RillIntakeClient {
    public constructor(private readonly config: RootConfig) {}

    public async fireEvent(event: MetricsEvent) {
        console.log("RillIntakeClient", event);
        // return axios.post(this.config.metrics.rillIntakeUrl, event);
    }
}
