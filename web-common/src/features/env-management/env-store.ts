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
      const entry =
        this.store.get(key) ??
        new EnvVariable(key, newEntries[key], this.version);
      entry.reconcile(newEntries[key], this.version);
      newStore.set(key, entry);
    }

    this.store = newStore;
  }

  public async flush(editSession: EnvEditSession) {
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
}
