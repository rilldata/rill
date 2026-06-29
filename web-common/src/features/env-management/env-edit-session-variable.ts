export class EnvEditSessionVariable {
  constructor(
    public readonly key: string,
    public value: string,
    public readonly mappedEnvVarName: string,
  ) {}
}
