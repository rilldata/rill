import { sidebarActions } from "@rilldata/web-common/features/chat/layouts/sidebar/sidebar-store.ts";
import { getMeasureDisplayName } from "@rilldata/web-common/features/dashboards/filters/getDisplayName.ts";
import { fetchExploreSpec } from "@rilldata/web-common/features/explores/selectors.ts";
import { prettyFormatTimeRange } from "@rilldata/web-common/lib/time/ranges/formatter.ts";
import { Interval } from "luxon";
import { get, writable } from "svelte/store";

export class MeasureSelection {
  public readonly measure = writable<string | null>(null);
  public readonly start = writable<Date | null>(null);
  public readonly end = writable<Date | null>(null);

  public setStart(measure: string, start: Date) {
    this.measure.set(measure);
    this.start.set(start);
    this.end.set(null);
  }

  public setRange(measure: string, start: Date, end: Date) {
    this.measure.set(measure);
    this.start.set(start);
    this.end.set(end);
  }

  public clear() {
    this.measure.set(null);
    this.start.set(null);
    this.end.set(null);
  }

  public hasSelection() {
    return Boolean(get(this.measure));
  }

  public async startAnomalyExplanationChat(
    instanceId: string,
    exploreName: string,
  ) {
    if (!this.hasSelection()) return;
    const measure = get(this.measure)!;

    const { metricsView } = await fetchExploreSpec(instanceId, exploreName);
    const metricsViewSpec = metricsView.metricsView?.state?.validSpec ?? {};

    const measureSpec = metricsViewSpec.measures?.find(
      (m) => m.name === measure,
    );
    if (!measureSpec) return;
    const measureDisplayName = getMeasureDisplayName(measureSpec);

    const start = get(this.start);
    const end = get(this.end);
    if (!start) return;
    let dataPointsPart = "";
    const formattedStart = prettyFormatTimeRange(
      Interval.fromDateTimes(start, start),
    );
    if (end) {
      const formattedEnd = prettyFormatTimeRange(
        Interval.fromDateTimes(end, end),
      );
      dataPointsPart = `data points in "${formattedStart} to ${formattedEnd}"`;
    } else {
      dataPointsPart = `data point on "${formattedStart}"`;
    }

    const prompt = `Please explain what drives the ${dataPointsPart} for "${measureDisplayName}". What dimensions have noticeably changed, as compared to other time windows?`;

    sidebarActions.startChat(instanceId, prompt);
  }
}

export const measureSelection = new MeasureSelection();
