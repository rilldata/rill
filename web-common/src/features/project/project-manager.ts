import type { V1GetProjectResponse } from "@rilldata/web-admin/client";
import { createQueryServiceClient } from "@rilldata/web-common/runtime-client/connectrpc";

class Project {
  projectData: V1GetProjectResponse;
  queryServiceClient: ReturnType<typeof createQueryServiceClient>;

  constructor(projectData: V1GetProjectResponse) {
    this.projectData = projectData;
    this.queryServiceClient = createQueryServiceClient(projectData);
  }
}

export function createDummyProject(
  organization: string = "default",
  project: string = "default",
): V1GetProjectResponse {
  return {
    project: {
      name: project,
      orgName: organization,
    },
    deployment: {
      runtimeInstanceId: "default",
      runtimeHost: "http://localhost:9009",
    },
  };
}

class ProjectManager {
  _map = new Map<string, Project>();

  addProject(projectData: V1GetProjectResponse) {
    const key = this._getProjectKey(
      projectData.project!.orgName!,
      projectData.project!.name!,
    );
    if (!this._map.has(key)) {
      const context = new Project(projectData);
      this._map.set(key, context);
    }
    return this._map.get(key)!;
  }

  getProjectContext({
    organization,
    project,
  }: {
    organization: string;
    project: string;
  }): Project {
    const key = this._getProjectKey(organization, project);

    return (
      this._map.get(key) ||
      this.addProject(createDummyProject(organization, project))
    );
  }

  private _getProjectKey(organization: string, project: string): string {
    return `${organization}::${project}`;
  }
}

export const projectManager = new ProjectManager();
