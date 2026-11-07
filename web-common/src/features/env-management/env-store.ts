import type { EnvEditSession } from "@rilldata/web-common/features/env-management/env-edit-session.ts";
import { EnvVariable } from "@rilldata/web-common/features/env-management/env-variable.ts";

export class EnvStore {
  public store = new Map<string, EnvVariable>();

  // Resolves once the first pull() has completed. Callers that allocate env var
  // names (the add-data forms) await this before constructing an edit session,
  // so the collision set is seeded from the persisted .env rather than an empty
  // store, which would otherwise let commit() overwrite an existing secret.
  public readonly ready: Promise<void>;
  private resolveReady!: () => void;
  private hasPulled = false;

  public constructor(
    private readonly getter: () => Promise<Record<string, string>>,
    private readonly setter: (entries: Record<string, string>) => Promise<void>,
  ) {
    this.ready = new Promise<void>((resolve) => {
      this.resolveReady = resolve;
    });
  }

  public whenReady() {
    return this.ready;
  }

  public async pull() {
    const newEntries = await this.getter();
    const newStore = new Map<string, EnvVariable>();
    for (const [key, value] of Object.entries(newEntries)) {
      newStore.set(key, new EnvVariable(key, value));
    }
    this.store = newStore;
    if (!this.hasPulled) {
      this.hasPulled = true;
      this.resolveReady();
    }
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
