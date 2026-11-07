import { EnvEditSessionVariable } from "@rilldata/web-common/features/env-management/env-edit-session-variable.ts";
import type { EnvStore } from "@rilldata/web-common/features/env-management/env-store.ts";
import { getName } from "@rilldata/web-common/features/entity-management/name-utils.ts";
import type { JSONSchemaObject } from "@rilldata/web-common/features/templates/schemas/types.ts";
import { getGenericEnvVarName } from "@rilldata/web-common/features/connectors/env-utils.ts";

export class EnvEditSession {
  public readonly entries = new Map<string, EnvEditSessionVariable>();
  public readonly inflightEntries = new Map<string, EnvEditSessionVariable>();

  // Final env var name → value this session has successfully committed.
  // Read by rollback to distinguish session-owned vars from externally
  // changed ones, and by startEdit to decide whether a prior allocation can
  // still be reused.
  private committed = new Map<string, string>();

  private assignedVars = new Set<string>();

  public constructor(
    public readonly parentStore: EnvStore,
    private readonly namespace: string = "",
    private readonly schema: JSONSchemaObject | null = null,
  ) {
    this.assignedVars = new Set<string>(this.parentStore.store.keys());
  }

  /**
   * Start a new edit session. Any prior allocations made by this session will be cleared,
   * and any externally changed values will be rolled back.
   * If changes are meant to be persisted, call commit().
   */
  public startEdit() {
    this.inflightEntries.clear();
    this.entries.forEach((v: EnvEditSessionVariable) => {
      const writtenValue = this.committed.get(v.mappedEnvVarName);
      if (writtenValue !== undefined) {
        // This entry was committed earlier in the session. If parent.store
        // no longer matches what we wrote, an external party has taken over
        // the name — drop our claim so the next acquire allocates a fresh
        // suffixed name instead of overwriting the external value.
        const current = this.parentStore.store.get(v.mappedEnvVarName);
        if (current?.value !== writtenValue) {
          this.committed.delete(v.mappedEnvVarName);
          return;
        }
      }
      this.inflightEntries.set(v.key, v);
    });
    this.entries.clear();

    // Seed the collision set with every live mapped name: both the parent
    // store's keys and the names still held by in-flight (uncommitted preview)
    // entries. Without the latter, a new acquire could hand out a name an
    // in-flight entry is about to reclaim, producing two entries with the same
    // mapped env var name.
    this.assignedVars = new Set<string>(this.parentStore.store.keys());
    this.inflightEntries.forEach((v) =>
      this.assignedVars.add(v.mappedEnvVarName),
    );
  }

  public acquire(key: string, value: string, envVarName?: string) {
    if (this.inflightEntries.has(key)) {
      const entry = this.inflightEntries.get(key)!;
      entry.value = value;
      this.inflightEntries.delete(key);
      this.entries.set(key, entry);
      this.assignedVars.add(entry.mappedEnvVarName);
      return entry;
    }

    envVarName ??= getGenericEnvVarName(this.namespace, key, this.schema);
    const mappedEnvVarName = this.acquireVarName(envVarName);
    const entry = new EnvEditSessionVariable(key, value, mappedEnvVarName);

    this.entries.set(key, entry);
    this.assignedVars.add(mappedEnvVarName);
    return entry;
  }

  public async commit() {
    await this.parentStore.flush(this);
    // Record successful writes so rollback can compare against them.
    this.entries.forEach((entry) => {
      this.committed.set(entry.mappedEnvVarName, entry.value);
    });
    this.inflightEntries.forEach((entry) => {
      this.committed.delete(entry.mappedEnvVarName);
    });
  }

  public async rollback() {
    await this.parentStore.rollbackOwned(this.committed);
    this.committed.clear();
  }

  private acquireVarName(varName: string) {
    const assignedVarName = getName(varName, [...this.assignedVars]);
    this.assignedVars.add(assignedVarName);
    return assignedVarName;
  }
}
