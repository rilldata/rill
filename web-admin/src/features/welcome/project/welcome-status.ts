import { get, type Writable } from "svelte/store";
import { sessionStorageStore } from "@rilldata/web-common/lib/store-utils/session-storage.ts";

const ProjectWelcomeStatusKey = "rill:welcome:project:status";

class ProjectWelcomeStatusStores {
  public isProjectWelcomeStep(project: string): boolean {
    const statusStore = this.get(project);
    return get(statusStore);
  }

  public setProjectWelcomeStep(project: string, value: boolean): void {
    const statusStore = this.get(project);
    statusStore.set(value);
  }

  private get(project: string): Writable<boolean> {
    const statusStore = sessionStorageStore(
      ProjectWelcomeStatusKey + ":" + project,
      false,
    );
    return statusStore;
  }
}

export const projectWelcomeStatusStores = new ProjectWelcomeStatusStores();

// Temporary localstorage based flag. Since our existing feature flag is at project level, we need separate flag.
const ProjectWelcomeEnabledKey = "rill:welcome:enabled";
export const projectWelcomeEnabled =
  localStorage.getItem(ProjectWelcomeEnabledKey) === "true";
