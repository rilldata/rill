import type { MetricsActionDefinition } from "$web-local/common/metrics-service/MetricsService";
import { fetchWrapper } from "../util/fetchWrapper";

export async function sendTelemetryEvent<
  EventType extends keyof MetricsActionDefinition
>(eventType: EventType, ...args: MetricsActionDefinition[EventType]) {
  await fetchWrapper(`v1/telemetry/${eventType}`, "POST", args as any);
}
