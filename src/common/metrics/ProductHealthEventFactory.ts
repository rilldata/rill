import { MetricsEventFactory } from "$common/metrics/MetricsEventFactory";
import type { ActiveEvent, CommonMetricsFields } from "$common/metrics/MetricsTypes";

export class ProductHealthEventFactory extends MetricsEventFactory {
    public activeEvent(commonMetricsInput: CommonMetricsFields,
                       durationMilSec: number, totalInFocus: number): ActiveEvent {
        const event = this.getBaseMetricsEvent(commonMetricsInput, "active") as ActiveEvent;
        event.duration_sec = Math.round(durationMilSec / 1000);
        event.total_in_focus = totalInFocus;
        return event;
    }
}
