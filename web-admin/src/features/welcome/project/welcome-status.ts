class ProjectWelcomeStatus {
  private projectWelcomeStatusMap = new Map<string, boolean>();

  public isProjectWelcomeStep(project: string): boolean {
    return this.projectWelcomeStatusMap.get(project) ?? false;
  }

  public setProjectWelcomeStep(project: string, value: boolean): void {
    this.projectWelcomeStatusMap.set(project, value);
  }
}

export const projectWelcomeStatus = new ProjectWelcomeStatus();

// Temporary localstorage based flag. Since our existing feature flag is at project level, we need separate flag.
const ProjectWelcomeEnabledKey = "rill:welcome:enabled";
export const projectWelcomeEnabled =
  localStorage.getItem(ProjectWelcomeEnabledKey) === "true";
