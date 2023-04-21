import type { V1DeploymentStatus } from "@rilldata/web-admin/client";
import type {
  RpcStatus,
  V1GetProjectResponse,
} from "@rilldata/web-admin/client";
import { getDashboardsForProject } from "@rilldata/web-admin/components/projects/dashboards";
import { invalidateProject } from "@rilldata/web-admin/components/projects/invalidations";
import type {
  QueryKey,
  QueryClient,
  CreateQueryResult,
} from "@tanstack/svelte-query";
import { get, Readable, Writable, writable } from "svelte/store";

export type ProjectStatusState = {
  orgName: string;
  projectName: string;
  queryRunning: boolean;
  pending: boolean;
  reconciling: boolean;
  errored: boolean;
  ready: boolean;
  prevStatus: V1DeploymentStatus;
};
const DefaultStatusValues = {
  queryRunning: false,
  pending: false,
  reconciling: false,
  errored: false,
  ready: false,
};

export const ProjectReconcilingPollTime = 1000; // 1 sec
export const ProjectErroredPollTime = 5000; // 5 sec
export const ProjectOkPollTime = 60 * 1000; // 1 min

export type ProjectStatusStore = Readable<ProjectStatusState>;

function setValueFromStatus(
  projectStatusStore: Writable<ProjectStatusState>,
  status: V1DeploymentStatus
) {
  projectStatusStore.update((state) => {
    // memoization
    if (state.prevStatus === status) return state;

    // set all flags to default to reset previous states
    for (const flag in DefaultStatusValues) {
      state[flag] = DefaultStatusValues[flag];
    }

    if (!status) {
      // query is still running
      state.queryRunning = true;
      return state;
    }

    switch (status) {
      case "DEPLOYMENT_STATUS_PENDING":
        state.pending = true;
        break;

      case "DEPLOYMENT_STATUS_RECONCILING":
        state.reconciling = true;
        break;

      case "DEPLOYMENT_STATUS_ERROR":
        state.errored = true;
        break;

      case "DEPLOYMENT_STATUS_OK":
        state.ready = true;
        break;
    }

    return state;
  });
}

async function projectIsReady(
  queryClient: QueryClient,
  projectStatusStore: Writable<ProjectStatusState>,
  projectData: V1GetProjectResponse
) {
  const projectStatusState = get(projectStatusStore);
  const dashboards = await getDashboardsForProject(projectData);
  // TODO: find a way to get the reconcile affected_paths and pass that instead of all dashboards
  return invalidateProject(
    queryClient,
    projectStatusState.orgName,
    projectStatusState.projectName,
    dashboards.map((dashboard) => dashboard.name)
  );
}

const stores = new Map<string, ProjectStatusStore>();

export function getProjectStatusStore(
  orgName: string,
  projectName: string,
  queryClient: QueryClient,
  getProjectQuery: CreateQueryResult<V1GetProjectResponse, RpcStatus> & {
    queryKey: QueryKey;
  },
  reconcilingPollTime = ProjectReconcilingPollTime,
  erroredPollTime = ProjectErroredPollTime,
  readyPollTime = ProjectOkPollTime
) {
  const key = `${orgName}__${projectName}`;
  if (stores.has(key)) return stores.get(key);

  const store = createProjectStatusStore(
    orgName,
    projectName,
    queryClient,
    getProjectQuery,
    reconcilingPollTime,
    erroredPollTime,
    readyPollTime
  );
  stores.set(key, store);
  return store;
}

export function createProjectStatusStore(
  orgName: string,
  projectName: string,
  queryClient: QueryClient,
  getProjectQuery: CreateQueryResult<V1GetProjectResponse, RpcStatus> & {
    queryKey: QueryKey;
  },
  reconcilingPollTime = ProjectReconcilingPollTime,
  erroredPollTime = ProjectErroredPollTime,
  readyPollTime = ProjectOkPollTime
): ProjectStatusStore {
  const projectStatusStore = writable<ProjectStatusState>({
    ...DefaultStatusValues,
    orgName,
    projectName,
    prevStatus: undefined,
  });

  getProjectQuery.subscribe((projectStatusResponse) => {
    const wasNotReady = !get(projectStatusStore).ready;

    setValueFromStatus(
      projectStatusStore,
      projectStatusResponse.data?.productionDeployment?.status
    );

    const state = get(projectStatusStore);
    let pollTime: number;
    if (state.reconciling) {
      pollTime = reconcilingPollTime;
    } else if (state.errored) {
      pollTime = erroredPollTime;
    } else if (state.ready) {
      pollTime = readyPollTime;
      if (wasNotReady) {
        projectIsReady(
          queryClient,
          projectStatusStore,
          get(getProjectQuery).data
        );
      }
    }

    if (pollTime) {
      queryClient.setQueryDefaults(getProjectQuery.queryKey, {
        refetchInterval: pollTime,
      });
    }
  });

  return projectStatusStore;
}
