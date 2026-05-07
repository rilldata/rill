import type { EnvVariable } from "@rilldata/web-common/features/env-management/env-variable.ts";

export class EnvEditSessionVariable {
  public mappedEnvVarName: string;
  public variable: EnvVariable | null = null;

  constructor(
    public readonly key: string,
    public value: string,
    public envVarName: string = key,
  ) {}
}
