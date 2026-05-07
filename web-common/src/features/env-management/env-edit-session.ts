import { EnvEditSessionVariable } from "@rilldata/web-common/features/env-management/env-edit-session-variable.ts";
import type { EnvStore } from "@rilldata/web-common/features/env-management/env-store.ts";
import { getName } from "@rilldata/web-common/features/entity-management/name-utils.ts";

export class EnvEditSession {
  public readonly entries = new Map<string, EnvEditSessionVariable>();
  public readonly inflightEntries = new Map<string, EnvEditSessionVariable>();

  private version = 0;
  private assignedVars = new Set<string>();

  public constructor(public readonly parentStore: EnvStore) {
    this.assignedVars = new Set<string>(this.parentStore.store.keys());
    this.version = this.parentStore.version;
  }

  public startEdit() {
    this.inflightEntries.clear();
    this.entries.forEach((v: EnvEditSessionVariable) => {
      if (v.variable && v.variable.version > this.version) return;
      this.inflightEntries.set(v.key, v);
    });
    this.entries.clear();

    this.assignedVars = new Set<string>(this.parentStore.store.keys());

    this.version = this.parentStore.version;
  }

  public acquire(key: string, value: string, envVarName?: string) {
    if (this.inflightEntries.has(key)) {
      const entry = this.inflightEntries.get(key)!;
      entry.value = value;
      this.inflightEntries.delete(key);
      this.entries.set(key, entry);
      this.assignedVars.add(key);
      return entry;
    }

    const entry = new EnvEditSessionVariable(key, value, envVarName);
    if (envVarName) entry.mappedEnvVarName = this.acquireVarName(envVarName);
    this.entries.set(key, entry);
    this.assignedVars.add(key);
    return entry;
  }

  public setValue(key: string, value: string) {
    const entry = this.entries.get(key);
    if (entry) {
      entry.value = value;
    }
  }

  public async commit() {
    await this.parentStore.flush(this);
  }

  private acquireVarName(varName: string) {
    const assignedVarName = getName(varName, [...this.assignedVars.keys()]);
    this.assignedVars.add(assignedVarName);
    return assignedVarName;
  }
}
