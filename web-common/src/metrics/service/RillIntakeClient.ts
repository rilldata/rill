import type { MetricsEvent } from "./MetricsTypes";

const RillIntakeUser = "data-modeler";
const RillIntakePassword =
  "lkh8T90ozWJP/KxWnQ81PexRzpdghPdzuB0ly2/86TeUU8q/bKiVug==";

export class RillIntakeClient {
  private readonly authHeader: string;

  public constructor(private readonly host: string) {
    // this is the format rill-intake expects.
    this.authHeader =
      "Basic " + btoa(`${RillIntakeUser}:${RillIntakePassword}`);
  }

  public async fireEvent(event: MetricsEvent) {
    if (!this.host) return;
    try {
      const resp = await fetch(this.host, {
        method: "POST",
        body: JSON.stringify(event),
        // headers: {
        //   Authorization: this.authHeader,
        // },
        credentials: "include",
      });
      if (!resp.ok)
        console.error(`Failed to send ${event.event_type}. ${resp.statusText}`);
    } catch (err) {
      console.error(`Failed to send ${event.event_type}. ${err.message}`);
    }
  }
}
