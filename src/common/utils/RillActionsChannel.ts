/**
 * A go like channel class that exposes a generator function.
 * One end will push redux actions after processing a request.
 * The other end will pop them and send it over an HTTP streaming connection.
 */
import { asyncWait, waitUntil } from "$common/utils/waitUtils";

type Message = {
  action: string;
  args: Array<any>;
};

export class RillActionsChannel {
  private messages = new Array<Message>();
  private isDone = false;

  public pushMessage(action: string, args: Array<any>) {
    this.messages.push({ action, args });
  }

  public end() {
    this.isDone = true;
  }

  public async *getActions(): AsyncGenerator {
    while (!this.isDone) {
      await waitUntil(() => this.messages.length > 0 || this.isDone, -1);
      while (this.messages.length > 0) {
        yield this.messages.pop();
      }
    }
  }
}
