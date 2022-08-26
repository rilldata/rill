import type { MetricsActionDefinition } from "$common/metrics-service/MetricsService";
import { fetchWrapper } from "$lib/util/fetchWrapper";

export async function sendTelemetryEvent<
  EventType extends keyof MetricsActionDefinition
>(eventType: EventType, ...args: MetricsActionDefinition[EventType]) {
  await fetchWrapper(`v1/telemetry/${eventType}`, "POST", args as any);
}
