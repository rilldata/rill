export class EnvVariable {
  constructor(
    public readonly key: string,
    public value: string,
    public version: number,
  ) {}

  public reconcile(newValue: string, version: number) {
    if (this.value === newValue) return;
    this.value = newValue;
    this.version = version;
  }
}
