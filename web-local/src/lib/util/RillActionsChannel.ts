/**
 * A go like channel class that exposes a generator function.
 * One end will push redux actions after processing a request.
 * The other end will pop them and send it over an HTTP streaming connection.
 */
import { waitUntil } from "./waitUtils";

export class RillActionsChannel {
  private messages = new Array<Record<string, unknown>>();
  private isDone = false;

  public pushMessage(rec: Record<string, unknown>) {
    this.messages.push(rec);
  }

  public end() {
    this.isDone = true;
  }

  public async *getActions(): AsyncGenerator<Record<string, unknown>> {
    while (!this.isDone) {
      await waitUntil(() => this.messages.length > 0 || this.isDone, -1);
      while (this.messages.length > 0) {
        yield this.messages.shift();
      }
    }
  }
}
