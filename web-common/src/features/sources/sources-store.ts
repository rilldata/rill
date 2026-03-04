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

export type SlowIngestionState =
  | null
  | { status: "loading"; filePath: string }
  | { status: "ingested"; filePath: string }
  | { status: "failed"; filePath: string; error: string };

class SourceIngestionTracker {
  private pending = new Set<string>();
  private loadingPaths = new Set<string>();
  public ingestedPath = writable<string | null>(null);
  public slowIngestion = writable<SlowIngestionState>(null);

  trackPending(filePath: string) {
    this.pending.add(filePath);
  }

  isPending(filePath: string) {
    return this.pending.has(filePath);
  }

  /** Import is taking long; show loading modal */
  trackLoading(filePath: string) {
    this.loadingPaths.add(filePath);
    this.slowIngestion.set({ status: "loading", filePath });
  }

  /** Ingestion finished; route to slow or fast path UI */
  trackIngested(filePath: string) {
    this.pending.delete(filePath);
    if (this.loadingPaths.has(filePath)) {
      this.loadingPaths.delete(filePath);
      this.slowIngestion.set({ status: "ingested", filePath });
    } else {
      this.ingestedPath.set(filePath);
    }
  }

  /** Import failed after loading modal was shown */
  trackFailed(filePath: string, error: string) {
    this.pending.delete(filePath);
    this.loadingPaths.delete(filePath);
    this.slowIngestion.set({ status: "failed", filePath, error });
  }

  /** Source creation was rolled back or failed (fast path) */
  trackCancelled(filePath: string) {
    this.pending.delete(filePath);
    this.loadingPaths.delete(filePath);
  }

  /** User dismissed a modal */
  dismiss() {
    this.ingestedPath.set(null);
    this.slowIngestion.set(null);
  }
}

export const sourceIngestionTracker = new SourceIngestionTracker();
