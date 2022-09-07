import axios from "axios";
import type { RootConfig } from "$common/config/RootConfig";
import type { MetricsEvent } from "$common/metrics-service/MetricsTypes";

export class RillIntakeClient {
  private readonly authHeader: string;

  public constructor(private readonly config: RootConfig) {
    // this is the format rill-intake expects.
    this.authHeader =
      "Basic " +
      Buffer.from(
        `${config.metrics.rillIntakeUser}:${config.metrics.rillIntakePassword}`
      ).toString("base64");
  }

  public async fireEvent(event: MetricsEvent) {
    // Debug Telemetry by uncommenting the below line
    // console.log(event);
    try {
      await axios.post(this.config.metrics.rillIntakeUrl, event, {
        headers: {
          Authorization: this.authHeader,
        },
      });
    } catch (err) {
      console.error(`Failed to send ${event.event_type}. ${err.message}`);
    }
  }
}
