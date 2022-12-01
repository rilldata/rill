import axios from "axios";
import type { RootConfig } from "@rilldata/web-local/common/config/RootConfig";
import type { MetricsEvent } from "./MetricsTypes";

export class RillIntakeClient {
  private readonly authHeader: string;

  public constructor(private readonly config: RootConfig) {
    // this is the format rill-intake expects.
    this.authHeader =
      "Basic " +
      btoa(
        `${config.metrics.rillIntakeUser}:${config.metrics.rillIntakePassword}`
      );
  }

  public async fireEvent(event: MetricsEvent) {
    try {
      await axios.post(`${RILL_RUNTIME_URL}/local/track`, event, {
        headers: {
          Authorization: this.authHeader,
        },
      });
    } catch (err) {
      console.error(`Failed to send ${event.event_type}. ${err.message}`);
    }
  }
}
