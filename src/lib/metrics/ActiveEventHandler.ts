import type { MetricsService } from "$common/metrics/MetricsService";

export class ActiveEventHandler {
    private isInFocus = true;
    private focusDuration = 0;
    private focusCount = 0;
    private previousInFocusTime = 0;

    public constructor(private metricsService: MetricsService) {
        window.addEventListener("blur", () => {
            console.log("lost focus");
            this.isInFocus = false;
            if (this.previousInFocusTime) {
                this.focusDuration += Date.now() - this.previousInFocusTime;
            }
        });
        window.addEventListener("focus", () => {
            console.log("gained focus");
            this.isInFocus = true;
            this.focusCount++;
            this.previousInFocusTime = Date.now();
        });

        // this is to ensure the event is triggered at the top of the minute
        setTimeout(() => {
            setInterval(() => {
                this.fireEvent();
            }, 60 * 1000);
        }, (60 - new Date().getSeconds()) * 1000);
    }

    private fireEvent() {
        if (this.focusCount === 0) return;
        if (this.previousInFocusTime) {
            this.focusDuration += Date.now() - this.previousInFocusTime;
        }

        this.metricsService.dispatch("activeEvent",
            [this.focusDuration, this.focusCount]);

        this.focusCount = 0;
        this.focusDuration = 0;
        this.previousInFocusTime = 0;
    }
}
