export class HttpStreamClient {
  public constructor(private readonly urlBase: string) {}

  public async request(
    path: string,
    method: string,
    data: Record<string, unknown>,
    messageCallback: (message: string) => void
  ): Promise<void> {
    const response = await fetch(`${this.urlBase}${path}`, {
      method,
      body: JSON.stringify(data),
      headers: { "Content-Type": "application/json" },
    });
    const reader = response.body.getReader();
    const decoder = new TextDecoder();

    let readResult = await reader.read();
    while (!readResult.done) {
      decoder
        .decode(readResult.value)
        .split("\x01")
        .forEach((message) => messageCallback(message));
      readResult = await reader.read();
    }
  }
}
