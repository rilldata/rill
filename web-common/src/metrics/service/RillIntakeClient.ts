import type { MetricsEvent } from "./MetricsTypes";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { get } from "svelte/store";

const RillIntakeUser = "data-modeler";
const RillIntakePassword =
  "lkh8T90ozWJP/KxWnQ81PexRzpdghPdzuB0ly2/86TeUU8q/bKiVug==";

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
    try {
      const resp = await fetch(`${get(runtime).host}/local/track`, {
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
