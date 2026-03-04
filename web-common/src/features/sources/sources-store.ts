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

export type IngestionState =
  | null
  | { status: "loading"; filePath: string }
  | { status: "ingested"; filePath: string }
  | { status: "failed"; filePath: string; error: string };

class SourceIngestionTracker {
  private pending = new Set<string>();
  private loadingPaths = new Set<string>();
  public ingestionState = writable<IngestionState>(null);

  trackPending(filePath: string) {
    this.pending.add(filePath);
  }

  isPending(filePath: string) {
    return this.pending.has(filePath);
  }

  /** Import is taking long; show loading modal */
  trackLoading(filePath: string) {
    this.loadingPaths.add(filePath);
    this.ingestionState.set({ status: "loading", filePath });
  }

  /** Ingestion finished; show success modal */
  trackIngested(filePath: string) {
    this.pending.delete(filePath);
    this.loadingPaths.delete(filePath);
    this.ingestionState.set({ status: "ingested", filePath });
  }

  /** Import failed after loading modal was shown */
  trackFailed(filePath: string, error: string) {
    this.pending.delete(filePath);
    this.loadingPaths.delete(filePath);
    this.ingestionState.set({ status: "failed", filePath, error });
  }

  /** Source creation was rolled back or failed before modal was shown */
  trackCancelled(filePath: string) {
    this.pending.delete(filePath);
    this.loadingPaths.delete(filePath);
  }

  /** User dismissed the modal */
  dismiss() {
    this.ingestionState.set(null);
  }
}

export const sourceIngestionTracker = new SourceIngestionTracker();
