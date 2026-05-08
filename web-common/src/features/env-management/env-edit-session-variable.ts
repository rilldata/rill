import type { EnvVariable } from "@rilldata/web-common/features/env-management/env-variable.ts";

export class EnvEditSessionVariable {
  public variable: EnvVariable | null = null;

  constructor(
    public readonly key: string,
    public value: string,
    public mappedEnvVarName: string = key,
  ) {}
}
