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
