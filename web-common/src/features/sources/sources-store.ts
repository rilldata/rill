import { writable, type Writable } from "svelte/store";

export enum DuplicateActions {
  None = "NONE",
  KeepBoth = "KEEP_BOTH",
  Overwrite = "OVERWRITE",
  Cancel = "CANCEL",
}

export const duplicateSourceAction: Writable<DuplicateActions> = writable(
  DuplicateActions.None,
);

export const duplicateSourceName: Writable<string | null> = writable(null);

class SourceIngestionTracker {
  private pending = new Set<string>();
  public ingestedPath = writable<string | null>(null);

  trackPending(filePath: string) {
    this.pending.add(filePath);
  }

  isPending(filePath: string) {
    return this.pending.has(filePath);
  }

  /** Ingestion finished — remove from pending and notify the UI */
  trackIngested(filePath: string) {
    this.pending.delete(filePath);
    this.ingestedPath.set(filePath);
  }

  /** Source creation was rolled back or failed */
  trackCancelled(filePath: string) {
    this.pending.delete(filePath);
  }

  /** User dismissed the success modal */
  dismiss() {
    this.ingestedPath.set(null);
  }
}

export const sourceIngestionTracker = new SourceIngestionTracker();
