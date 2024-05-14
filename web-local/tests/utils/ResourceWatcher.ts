import { waitUntil } from "@rilldata/web-common/lib/waitUtils";
import type { Page } from "playwright";
import { updateCodeEditor } from "web-local/tests/utils/commonHelpers";

const ResourceWatcherLogRegex = /^\[(.*)] rill\.runtime\.v1\.(.*)\/(.*)$/;

export class ResourceWatcher {
  private statuses = new Map<string, string>();

  public constructor(private readonly page: Page) {
    page.on("console", (e) => {
      const matches = ResourceWatcherLogRegex.exec(e.text());
      if (!matches || matches.length < 4) return;
      const [, status, , name] = matches;
      this.statuses.set(name, status);
    });
  }

  public async waitForResource(name: string) {
    return waitUntil(() => this.statuses.get(name) === "RECONCILE_STATUS_IDLE");
  }

  public async updateAndWaitForDashboard(
    code: string,
    name = "AdBids_model_dashboard",
  ) {
    this.statuses.delete(name); // clear older state
    return Promise.all([
      updateCodeEditor(this.page, code),
      this.waitForResource(name),
    ]);
  }
}
