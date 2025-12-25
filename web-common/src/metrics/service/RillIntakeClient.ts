import httpClient from "@rilldata/web-common/runtime-client/http-client";
import type { MetricsEvent } from "./MetricsTypes";

const RillIntakeUser = import.meta.env.RILL_UI_PUBLIC_INTAKE_USER;
const RillIntakePassword = import.meta.env.RILL_UI_PUBLIC_INTAKE_PASSWORD;

export interface TelemetryClient {
  fireEvent(event: MetricsEvent): Promise<void>;
}

export class RillIntakeClient implements TelemetryClient {
  private readonly authHeader: string;

  public constructor() {
    // this is the format rill-intake expects.
    this.authHeader =
      "Basic " + btoa(`${RillIntakeUser}:${RillIntakePassword}`);
  }

  public async fireEvent(event: MetricsEvent) {
    if (!RillIntakeUser || !RillIntakePassword) return;

    try {
      const resp = await fetch(`${httpClient.getHost()}/local/track`, {
        method: "POST",
        body: JSON.stringify(event),
        headers: {
          Authorization: this.authHeader,
        },
      });
      if (!resp.ok)
        console.error(`Failed to send ${event.event_type}. ${resp.statusText}`);
    } catch (err) {
      console.error(`Failed to send ${event.event_type}. ${err.message}`);
    }
  }
}
