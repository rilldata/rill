import { page } from "$app/stores";
import { getScreenNameFromPage } from "@rilldata/web-admin/features/navigation/nav-utils";
import { RillAdminTelemetryClient } from "@rilldata/web-admin/features/telemetry/RillAdminTelemetryClient";
import { BehaviourEventHandler } from "@rilldata/web-common/metrics/BehaviourEventHandler";
import { collectCommonUserFields } from "@rilldata/web-common/metrics/collectCommonUserFields";
import { ErrorEventHandler } from "@rilldata/web-common/metrics/ErrorEventHandler";
import {
  setBehaviourEvent,
  setErrorEvent,
  setMetricsService,
} from "@rilldata/web-common/metrics/initMetrics";
import { BehaviourEventFactory } from "@rilldata/web-common/metrics/service/BehaviourEventFactory";
import { ErrorEventFactory } from "@rilldata/web-common/metrics/service/ErrorEventFactory";
import { MetricsService } from "@rilldata/web-common/metrics/service/MetricsService";
import { ProductHealthEventFactory } from "@rilldata/web-common/metrics/service/ProductHealthEventFactory";
import { onDestroy } from "svelte";
import { get } from "svelte/store";

export const cloudVersion = import.meta.env.RILL_UI_PUBLIC_VERSION;

export async function initCloudMetrics() {
  const telemetryClient = new RillAdminTelemetryClient();

  const metricsService = new MetricsService(telemetryClient, [
    new ProductHealthEventFactory(),
    new BehaviourEventFactory(),
    new ErrorEventFactory(),
  ]);
  setMetricsService(metricsService);

  // --- Flush telemetry on unload/visibilitychange ---
  const flushTelemetry = () => {
    telemetryClient.flush(true);
  };
  window.addEventListener("beforeunload", flushTelemetry);
  document.addEventListener("visibilitychange", () => {
    if (document.visibilityState === "hidden") flushTelemetry();
  });
  onDestroy(() => {
    window.removeEventListener("beforeunload", flushTelemetry);
    document.removeEventListener("visibilitychange", flushTelemetry);
  });

  const commonUserMetrics = await collectCommonUserFields();
  setBehaviourEvent(
    new BehaviourEventHandler(metricsService, commonUserMetrics),
  );
  setErrorEvent(
    new ErrorEventHandler(
      metricsService,
      commonUserMetrics,
      window.location.host.startsWith("localhost"),
      () => getScreenNameFromPage(get(page)),
    ),
  );
  // TODO: add other handlers and callers
}
