import type { EnvEditSession } from "@rilldata/web-common/features/env-management/env-edit-session.ts";
import { EnvVariable } from "@rilldata/web-common/features/env-management/env-variable.ts";

export class EnvStore {
  public store = new Map<string, EnvVariable>();
  public version = 0;

  public constructor(
    private readonly getter: () => Promise<Record<string, string>>,
    private readonly setter: (entries: Record<string, string>) => Promise<void>,
  ) {}

  public async pull() {
    this.version++;
    const newEntries = await this.getter();
    const newStore = new Map<string, EnvVariable>();

    for (const key in newEntries) {
      let entry: EnvVariable;
      if (this.store.has(key)) {
        entry = this.store.get(key)!;
        entry.reconcile(newEntries[key], this.version);
      } else {
        entry = new EnvVariable(key, newEntries[key], this.version);
      }
      newStore.set(key, entry);
    }

    this.store = newStore;
  }

  public async flush(editSession: EnvEditSession) {
    if (
      editSession.entries.size === 0 &&
      editSession.inflightEntries.size === 0
    ) {
      return;
    }

    editSession.entries.forEach((entry) => {
      if (!entry.variable) {
        entry.variable = new EnvVariable(
          entry.mappedEnvVarName,
          entry.value,
          this.version,
        );
      } else {
        entry.variable.value = entry.value;
      }
      // Use mappedEnvVarName to add the final entry, key could be yaml key in edit session.
      this.store.set(entry.mappedEnvVarName, entry.variable);
    });
    editSession.inflightEntries.forEach((entry) => {
      this.store.delete(entry.mappedEnvVarName);
    });

    await this.setter(
      Object.fromEntries(this.store.values().map((v) => [v.key, v.value])),
    );
  }

  public async rollback(editSession: EnvEditSession) {
    if (
      editSession.entries.size === 0 &&
      editSession.inflightEntries.size === 0
    ) {
      return;
    }

    editSession.entries.forEach((entry) => {
      this.store.delete(entry.mappedEnvVarName);
    });
    editSession.inflightEntries.forEach((entry) => {
      this.store.delete(entry.mappedEnvVarName);
    });

    await this.setter(
      Object.fromEntries(this.store.values().map((v) => [v.key, v.value])),
    );
  }
}
