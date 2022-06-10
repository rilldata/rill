export class RillDeveloperClient {
  public async *sendRequest(action: string, args: Array<any>) {
    const resp = await fetch("/api?action=" + action);
  }
}
