import { type Readable, type Updater, writable } from "svelte/store";

export enum ChartPromptStatus {
  Idle,
  GeneratingData,
  GeneratingChartSpec,
}
export type ChartPromptState = {
  charts: Record<string, ChartPromptStatus>;
};
export type ChartPromptStore = Readable<ChartPromptState> &
  ReturnType<typeof chartPromptStoreActions>;

function chartPromptStoreActions(
  update: (this: void, updater: Updater<ChartPromptState>) => void,
) {
  return {
    setStatus(chartPath: string, status: ChartPromptStatus) {
      update((s) => {
        s.charts[chartPath] = status;
        return s;
      });
    },

    deleteStatus(chartPath: string) {
      update((s) => {
        delete s.charts[chartPath];
        return s;
      });
    },
  };
}

function createChartPromptStore(): ChartPromptStore {
  const { subscribe, update } = writable<ChartPromptState>({
    charts: {},
  });

  return {
    subscribe,
    ...chartPromptStoreActions(update),
  };
}

export const chartPromptStore = createChartPromptStore();
