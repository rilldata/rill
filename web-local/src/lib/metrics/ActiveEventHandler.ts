import type { MetricsService } from "@rilldata/web-local/lib/metrics/service/MetricsService";
import type { CommonUserFields } from "@rilldata/web-local/lib/metrics/service/MetricsTypes";

const ActiveEventInterval = 60;

export class ActiveEventHandler {
  private isInFocus = true;
  private focusDuration = 0;
  private focusCount = 0;
  private previousInFocusTime = 0;

  public constructor(
    private readonly metricsService: MetricsService,
    private readonly commonUserMetrics: CommonUserFields
  ) {
    window.addEventListener("blur", () => {
      this.isInFocus = false;
      if (this.previousInFocusTime) {
        this.focusDuration += Date.now() - this.previousInFocusTime;
      }
      this.previousInFocusTime = 0;
    });
    window.addEventListener("focus", () => {
      this.isInFocus = true;
      this.focusCount++;
      this.previousInFocusTime = Date.now();
    });

    // this is to ensure the event is triggered at the top of the minute
    setTimeout(() => {
      setInterval(() => {
        this.fireEvent();
      }, ActiveEventInterval * 1000);
    }, (60 - new Date().getSeconds()) * 1000);
  }

  private fireEvent() {
    if (this.previousInFocusTime) {
      this.focusDuration += Date.now() - this.previousInFocusTime;
    }

    if (this.focusCount > 0) {
      return this.metricsService.dispatch("activeEvent", [
        this.commonUserMetrics,
        this.focusDuration,
        this.focusCount,
      ]);
    }

    this.focusCount = 0;
    this.focusDuration = 0;
    this.previousInFocusTime = Date.now();
  }
}
