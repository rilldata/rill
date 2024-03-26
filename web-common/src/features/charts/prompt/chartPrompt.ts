import { localStorageStore } from "@rilldata/web-common/lib/store-utils";
import { derived, get, type Readable, writable } from "svelte/store";

export enum ChartPromptStatus {
  Idle,
  GeneratingData,
  GeneratingChartSpec,
  Error,
}

export type ChartPrompt = {
  entityName: string;
  chartName: string;
  prompt: string;
  time: number;
  status: ChartPromptStatus;
  error?: string;
};

export type ChartPromptHistory = {
  entities: Record<string, Array<ChartPrompt>>;
  count: number;
};

export class ChartPromptsStore {
  private readonly history = localStorageStore<ChartPromptHistory>(
    "chart-prompt",
    {
      entities: {},
      count: 0,
    },
  );

  private readonly chartStatus = writable<Record<string, ChartPrompt>>({});

  public constructor(private readonly maxHistoryCount: number) {}

  public getStatusForChart(
    chartName: string,
  ): Readable<ChartPrompt | undefined> {
    return derived(this.chartStatus, (cs) => cs[chartName]);
  }

  public getHistoryForEntity(entityName: string): Readable<Array<ChartPrompt>> {
    return derived(this.history, (h) => h.entities[entityName] ?? []);
  }

  public startPrompt(entityName: string, chartName: string, prompt: string) {
    const chartPrompt: ChartPrompt = {
      entityName,
      chartName,
      prompt,
      time: Date.now(),
      status:
        entityName === chartName
          ? ChartPromptStatus.GeneratingData
          : ChartPromptStatus.GeneratingChartSpec,
    };
    this.chartStatus.update((cs) => {
      cs[chartName] = chartPrompt;
      return cs;
    });
    this.addToHistory(entityName, chartPrompt);
  }

  public updatePromptStatus(chartName: string, status: ChartPromptStatus) {
    this.chartStatus.update((cs) => {
      if (cs[chartName]) cs[chartName].status = status;
      return cs;
    });
    this.history.update((h) => {
      if (!h.entities[chartName]?.length) return;
      h.entities[chartName][0].status = status;
    });
  }

  public setPromptError(chartName: string, error: string) {
    this.chartStatus.update((cs) => {
      if (cs[chartName]) {
        cs[chartName].status = ChartPromptStatus.Error;
        cs[chartName].error = error;
      }
      return cs;
    });
    this.history.update((h) => {
      if (!h.entities[chartName]?.length) return;
      h.entities[chartName][0].status = ChartPromptStatus.Error;
    });
  }

  private addToHistory(entityName: string, newEntry: ChartPrompt) {
    let history = get(this.history);
    history.entities[entityName] ??= [];

    const existingPromptIdx = history.entities[entityName].findIndex(
      (p) => p.prompt === newEntry.prompt,
    );
    if (existingPromptIdx >= 0) {
      history.entities[entityName].splice(existingPromptIdx, 1);
    }
    history.entities[entityName].unshift(newEntry);

    while (history.count >= this.maxHistoryCount) {
      history = this.removeOldestHistoryEvent(history);
    }

    this.history.set(history);
  }

  private removeOldestHistoryEvent(history: ChartPromptHistory) {
    let oldestPrompt: ChartPrompt | undefined;
    let oldestEntity: string | undefined;

    for (const entity in history) {
      if (history.entities[entity].length === 0) continue;
      const prompt =
        history.entities[entity][history.entities[entity].length - 1];
      if (!oldestPrompt || prompt.time < oldestPrompt.time) {
        oldestPrompt = prompt;
        oldestEntity = entity;
      }
    }

    if (oldestEntity) {
      history.entities[oldestEntity].pop();
      history.count--;
    }

    return history;
  }
}

export const chartPromptsStore = new ChartPromptsStore(10);
