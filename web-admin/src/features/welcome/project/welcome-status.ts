import { get, type Writable } from "svelte/store";
import { sessionStorageStore } from "@rilldata/web-common/lib/store-utils/session-storage.ts";

const ProjectWelcomeStatusKey = "rill:welcome:project:status";

class ProjectWelcomeStatusStores {
  private welcomeStatus = new Map<string, Writable<boolean>>();

  public inProjectWelcomeStep(project: string): boolean {
    const statusStore = this.get(project);
    return get(statusStore);
  }

  public setProjectWelcomeStatus(project: string, value: boolean): void {
    const statusStore = this.get(project);
    statusStore.set(value);
  }

  private get(project: string): Writable<boolean> {
    if (this.welcomeStatus.has(project)) {
      return this.welcomeStatus.get(project);
    }
    const statusStore = sessionStorageStore(
      ProjectWelcomeStatusKey + ":" + project,
      false,
    );
    this.welcomeStatus.set(project, statusStore);
    return statusStore;
  }
}

export const projectWelcomeStatusStores = new ProjectWelcomeStatusStores();
