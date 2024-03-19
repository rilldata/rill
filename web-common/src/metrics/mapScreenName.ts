import { MetricsEventScreenName } from "@rilldata/web-common/metrics/service/MetricsTypes";
import type { Page } from "@sveltejs/kit";

/**
 * mapScreenName maps page route id to {@link MetricsEventScreenName}
 * This is only for web-local. But since {@link initMetrics} is in web-common this has to be here for now.
 */
export function mapScreenName(page: Page): MetricsEventScreenName {
  switch (page.route.id) {
    case "/(application)/dashboard/[name]":
      return MetricsEventScreenName.Dashboard;
    case "/(application)/dashboard/[name]/edit":
      return MetricsEventScreenName.MetricsDefinition;
    case "/(application)/source/[name]":
      return MetricsEventScreenName.Source;
    case "/(application)/model/[name]":
      return MetricsEventScreenName.Model;
    case "/(application)/chart/[name]":
      return MetricsEventScreenName.Chart;
    case "/(application)/custom-dashboard/[name]":
      return MetricsEventScreenName.CustomDashboard;
    case "/(application)":
      return MetricsEventScreenName.Home;
  }
  return MetricsEventScreenName.Unknown;
}
