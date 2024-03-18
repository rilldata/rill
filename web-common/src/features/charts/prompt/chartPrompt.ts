import { localStorageStore } from "@rilldata/web-common/lib/store-utils";
import { derived, get, type Readable } from "svelte/store";

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
  // TODO: store error as well?
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

  public constructor(private readonly maxHistoryCount: number) {}

  public getStatusForChart(
    chartName: string,
  ): Readable<ChartPrompt | undefined> {
    return derived(this.history, (h) => h.entities[chartName]?.[0]);
  }

  public getHistoryForEntity(entityName: string): Readable<Array<ChartPrompt>> {
    return derived(this.history, (h) => h.entities[entityName] ?? []);
  }

  public startPrompt(entityName: string, chartName: string, prompt: string) {
    this.addToHistory(entityName, {
      entityName,
      chartName,
      prompt,
      time: Date.now(),
      status:
        entityName === chartName
          ? ChartPromptStatus.GeneratingData
          : ChartPromptStatus.GeneratingChartSpec,
    });
  }

  public updatePromptStatus(chartName: string, status: ChartPromptStatus) {
    this.history.update((h) => {
      if (!h.entities[chartName]?.length) return;
      h.entities[chartName][0].status = status;
    });
  }

  private addToHistory(entityName: string, newEntry: ChartPrompt) {
    let history = get(this.history);
    history.entities[entityName] ??= [];
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
