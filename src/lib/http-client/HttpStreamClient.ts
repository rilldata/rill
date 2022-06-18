import type { Message } from "$common/utils/RillActionsChannel";
import { ReduxActions } from "$lib/redux-store/ActionTypes";

export class HttpStreamClient {
  public static instance: HttpStreamClient;

  public constructor(
    private readonly urlBase: string,
    private readonly dispatch: (action: any) => void
  ) {}

  public async request(
    path: string,
    method: string,
    data: Record<string, unknown> = undefined
  ): Promise<void> {
    const response = await fetch(`${this.urlBase}${path}`, {
      method,
      ...(data ? { body: JSON.stringify(data) } : {}),
      headers: { "Content-Type": "application/json" },
    });
    const reader = response.body.getReader();
    const decoder = new TextDecoder();

    let readResult = await reader.read();
    while (!readResult.done) {
      decoder
        .decode(readResult.value)
        .split("\x01")
        .forEach((message) => this.dispatchMessage(message));
      readResult = await reader.read();
    }
  }

  public static create(urlBase: string, dispatch: (action: any) => void) {
    this.instance = new HttpStreamClient(urlBase, dispatch);
  }

  private dispatchMessage(message: string) {
    if (!message) return;
    try {
      const messageJson: Message<any> = JSON.parse(message);
      this.dispatch(
        ReduxActions[messageJson.action].apply(null, messageJson.args)
      );
    } catch (err) {
      console.error(err);
    }
  }
}
