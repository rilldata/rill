import type { EnvEditSession } from "@rilldata/web-common/features/env-management/env-edit-session.ts";
import { EnvVariable } from "@rilldata/web-common/features/env-management/env-variable.ts";

export class EnvStore {
  public store = new Map<string, EnvVariable>();

  public constructor(
    private readonly getter: () => Promise<Record<string, string>>,
    private readonly setter: (entries: Record<string, string>) => Promise<void>,
  ) {}

  public async pull() {
    const newEntries = await this.getter();
    const newStore = new Map<string, EnvVariable>();
    for (const key in newEntries) {
      newStore.set(key, new EnvVariable(key, newEntries[key]));
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

    const newStore = new Map<string, EnvVariable>(this.store.entries());

    editSession.entries.forEach((entry) => {
      newStore.set(
        entry.mappedEnvVarName,
        new EnvVariable(entry.mappedEnvVarName, entry.value),
      );
    });
    editSession.inflightEntries.forEach((entry) => {
      newStore.delete(entry.mappedEnvVarName);
    });

    await this.setter(
      Object.fromEntries([...newStore.values()].map((v) => [v.key, v.value])),
    );

    this.store = newStore;
  }

  // Remove only the vars whose current value still matches what the session
  // wrote. Anything an external party has changed (or removed) since the
  // commit is left alone — we only revert our own writes.
  public async rollbackOwned(committed: Map<string, string>) {
    if (committed.size === 0) return;

    const newStore = new Map<string, EnvVariable>(this.store.entries());
    let changed = false;
    for (const [name, writtenValue] of committed) {
      const current = newStore.get(name);
      if (current?.value === writtenValue) {
        newStore.delete(name);
        changed = true;
      }
    }
    if (!changed) return;

    await this.setter(
      Object.fromEntries([...newStore.values()].map((v) => [v.key, v.value])),
    );

    this.store = newStore;
  }
}
