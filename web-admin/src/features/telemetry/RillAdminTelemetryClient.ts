import { ADMIN_URL } from "@rilldata/web-admin/client/http-client";
import type { MetricsEvent } from "@rilldata/web-common/metrics/service/MetricsTypes";
import type { TelemetryClient } from "@rilldata/web-common/metrics/service/RillIntakeClient";

const BATCH_SIZE = 10;
const BATCH_TIMEOUT_MS = 30 * 1000; // 30 seconds

export class RillAdminTelemetryClient implements TelemetryClient {
  private eventQueue: MetricsEvent[] = [];
  private batchTimer: ReturnType<typeof setTimeout> | null = null;

  public async fireEvent(event: MetricsEvent) {
    this.eventQueue.push(event);
    if (this.eventQueue.length >= BATCH_SIZE) {
      this.sendBatch();
    } else if (!this.batchTimer) {
      this.batchTimer = setTimeout(() => {
        this.sendBatch();
      }, BATCH_TIMEOUT_MS);
    }
  }

  /**
   * Flushes all queued events. If useBeacon is true, attempts to use navigator.sendBeacon for delivery (for unload scenarios).
   */
  public async flush(useBeacon: boolean = false) {
    if (this.eventQueue.length > 0) {
      await this.sendBatch(useBeacon);
    }
  }

  /**
   * Sends the current batch of events. If useBeacon is true and available, uses navigator.sendBeacon for delivery.
   */
  private async sendBatch(useBeacon: boolean = false) {
    if (this.batchTimer) {
      clearTimeout(this.batchTimer);
      this.batchTimer = null;
    }
    if (this.eventQueue.length === 0) return;
    const eventsToSend = this.eventQueue;
    this.eventQueue = [];

    try {
      if (useBeacon && this._sendWithBeacon(eventsToSend)) {
        // Sent with beacon, nothing more to do.
        return;
      }
      await this._sendWithFetch(eventsToSend);
    } catch (err: any) {
      console.error(`Failed to send batch of events. ${err.message}`);
    }
  }

  /**
   * Try to send events with sendBeacon. Returns true if attempted.
   *
   * Why use sendBeacon for unloads instead of fetch?
   * ------------------------------------------------
   * Browsers do not guarantee that asynchronous operations (like fetch) will complete
   * when a page is unloading (e.g., tab close, refresh, navigation). As a result, telemetry
   * or analytics events sent with fetch may be lost if the page unloads before the request finishes.
   *
   * The navigator.sendBeacon API is specifically designed for this scenario: it allows you to
   * reliably send small amounts of data to a server during page unload. The browser will attempt
   * to deliver the data in the background, even as the page is closing, making it the recommended
   * approach for sending telemetry on unload events.
   */
  private _sendWithBeacon(events: MetricsEvent[]): boolean {
    if (typeof navigator !== "undefined" && navigator.sendBeacon) {
      const blob = new Blob([JSON.stringify({ events })], {
        type: "application/json",
      });
      const ok = navigator.sendBeacon(`${ADMIN_URL}/v1/telemetry/events`, blob);
      if (!ok) {
        console.error("sendBeacon failed to queue telemetry events.");
      }
      return true;
    }
    return false;
  }

  /** Send events with fetch. */
  private async _sendWithFetch(events: MetricsEvent[]) {
    const resp = await fetch(`${ADMIN_URL}/v1/telemetry/events`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ events }),
      credentials: "include",
    });
    if (!resp.ok) {
      console.error(`Failed to send batch of events. ${resp.statusText}`);
    }
  }
}
