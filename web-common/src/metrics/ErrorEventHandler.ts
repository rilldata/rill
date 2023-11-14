import { page } from "$app/stores";
import type { MetricsService } from "@rilldata/web-common/metrics/service/MetricsService";
import type {
  CommonUserFields,
  MetricsEventSpace,
} from "@rilldata/web-common/metrics/service/MetricsTypes";
import { MetricsEventScreenName } from "@rilldata/web-common/metrics/service/MetricsTypes";
import { get } from "svelte/store";
import type {
  SourceConnectionType,
  SourceErrorCodes,
  SourceFileType,
} from "./service/SourceEventTypes";

export class ErrorEventHandler {
  public constructor(
    private readonly metricsService: MetricsService,
    private readonly commonUserMetrics: CommonUserFields
  ) {
    this.commonUserMetrics = commonUserMetrics;
  }

  public fireSourceErrorEvent(
    space: MetricsEventSpace,
    screen_name: MetricsEventScreenName,
    error_code: SourceErrorCodes,
    connection_type: SourceConnectionType,
    file_type: SourceFileType,
    glob: boolean
  ) {
    return this.metricsService.dispatch("sourceErrorEvent", [
      this.commonUserMetrics,
      space,
      screen_name,
      error_code,
      connection_type,
      file_type,
      glob,
    ]);
  }

  public fireErrorBoundaryEvent(api: string, status: string, message: string) {
    return this.metricsService.dispatch("errorBoundaryEvent", [
      this.commonUserMetrics,
      this.getScreenNameFromPage(),
      api,
      status,
      message,
    ]);
  }

  private getScreenNameFromPage() {
    switch (get(page).route.id) {
      case "/[organization]/[project]":
        return MetricsEventScreenName.Project;
      case "/[organization]/[project]/[dashboard]":
        return MetricsEventScreenName.Dashboard;
      case "/[organization]/[project]/[dashboard]/-/reports/[report]":
        return MetricsEventScreenName.Report;
      case "/[organization]/[project]/[dashboard]/-/reports/[report]/export":
        return MetricsEventScreenName.ReportExport;
      default:
        return MetricsEventScreenName.Home; // acts a catch-all for now
    }
  }
}
