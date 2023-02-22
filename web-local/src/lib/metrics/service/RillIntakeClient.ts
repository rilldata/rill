import { fetchWrapper } from "@rilldata/web-local/lib/util/fetchWrapper";
import type { MetricsEvent } from "./MetricsTypes";

const RillIntakeUser = "data-modeler";
const RillIntakePassword =
  "lkh8T90ozWJP/KxWnQ81PexRzpdghPdzuB0ly2/86TeUU8q/bKiVug==";

export class RillIntakeClient {
  private readonly authHeader: string;

  public constructor() {
    // this is the format rill-intake expects.
    this.authHeader =
      "Basic " + btoa(`${RillIntakeUser}:${RillIntakePassword}`);
  }

  public async fireEvent(event: MetricsEvent) {
    try {
      const resp = await fetch(`${RILL_RUNTIME_URL}/local/track`, {
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
