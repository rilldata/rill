import type { V1Log } from "@rilldata/web-common/runtime-client";

export type LogEntry = V1Log & { _id: number };

export interface LogFilters {
  levels?: string[];
  search?: string;
}

/**
 * In-memory log buffer for ProjectLogsPage. Keeps the state machine out of
 * the Svelte component so the filtering + ring-buffer behavior can be
 * tested without a DOM.
 */
export class ProjectLogsStore {
  private readonly entries: LogEntry[] = [];
  private nextId = 0;

  constructor(public readonly maxLogs: number) {}

  public addLog(log: V1Log): LogEntry {
    const entry: LogEntry = { ...log, _id: this.nextId++ };
    this.entries.push(entry);
    if (this.entries.length > this.maxLogs) {
      this.entries.shift();
    }
    return entry;
  }

  public getAll(): LogEntry[] {
    return [...this.entries];
  }

  public getFiltered(filters: LogFilters = {}): LogEntry[] {
    const { levels = [], search = "" } = filters;
    const needle = search.toLowerCase();
    return this.entries.filter((log) => {
      const matchesLevel =
        levels.length === 0 || levels.includes(log.level ?? "");
      if (!matchesLevel) return false;

      if (!needle) return true;
      return (
        (log.message?.toLowerCase().includes(needle) ?? false) ||
        (log.jsonPayload?.toLowerCase().includes(needle) ?? false)
      );
    });
  }

  public clear(): void {
    this.entries.length = 0;
  }
}
