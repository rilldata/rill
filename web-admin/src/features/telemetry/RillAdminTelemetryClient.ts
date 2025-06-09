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

  public async flush() {
    if (this.eventQueue.length > 0) {
      await this.sendBatch();
    }
  }

  private async sendBatch() {
    if (this.batchTimer) {
      clearTimeout(this.batchTimer);
      this.batchTimer = null;
    }
    if (this.eventQueue.length === 0) return;
    const eventsToSend = this.eventQueue;
    this.eventQueue = [];
    try {
      const resp = await fetch(`${ADMIN_URL}/v1/telemetry/events`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          events: eventsToSend,
        }),
        credentials: "include",
      });
      if (!resp.ok)
        console.error(`Failed to send batch of events. ${resp.statusText}`);
    } catch (err: any) {
      console.error(`Failed to send batch of events. ${err.message}`);
    }
  }
}
