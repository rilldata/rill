/**
 * A go like channel class that exposes a generator function.
 * One end will push redux actions after processing a request.
 * The other end will pop them and send it over an HTTP streaming connection.
 */
import { waitUntil } from "$common/utils/waitUtils";
import type { ReduxActionDefinitions } from "$lib/redux-store/ActionTypes";

export type Message<Action extends keyof ReduxActionDefinitions> = {
  action: Action;
  args: ReduxActionDefinitions[Action];
};

export class RillActionsChannel {
  private messages = new Array<Message<any>>();
  private isDone = false;

  public pushMessage<Action extends keyof ReduxActionDefinitions>(
    action: Action,
    args: ReduxActionDefinitions[Action]
  ) {
    this.messages.push({ action, args });
  }

  public end() {
    this.isDone = true;
  }

  public async *getActions(): AsyncGenerator<Message<any>> {
    while (!this.isDone) {
      await waitUntil(() => this.messages.length > 0 || this.isDone, -1);
      while (this.messages.length > 0) {
        yield this.messages.pop();
      }
    }
  }
}
