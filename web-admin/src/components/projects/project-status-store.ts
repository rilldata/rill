import type { V1DeploymentStatus } from "@rilldata/web-admin/client";
import type {
  RpcStatus,
  V1GetProjectResponse,
} from "@rilldata/web-admin/client";
import type {
  QueryKey,
  QueryClient,
  CreateQueryResult,
} from "@tanstack/svelte-query";
import { get, Readable, Writable, writable } from "svelte/store";

export type ProjectStatusState = {
  queryRunning: boolean;
  pending: boolean;
  reconciling: boolean;
  errored: boolean;
  ok: boolean;
  prevStatus: V1DeploymentStatus;
};
const DefaultStatusValues: ProjectStatusState = {
  queryRunning: false,
  pending: false,
  reconciling: false,
  errored: false,
  ok: false,
  prevStatus: undefined,
};

export const ProjectReconcilingPollTime = 1000; // 1 sec
export const ProjectErroredPollTime = 5000; // 5 sec
export const ProjectOkPollTime = 60 * 1000; // 1 min

export type ProjectStatusStore = Readable<ProjectStatusState>;

function setValueFromStatus(
  projectStatusStore: Writable<ProjectStatusState>,
  status: V1DeploymentStatus
) {
  // memoization
  if (get(projectStatusStore).prevStatus === status) return;

  if (!status) {
    // query is still running
    projectStatusStore.set({
      ...DefaultStatusValues,
      queryRunning: true,
      prevStatus: status,
    });
    return;
  }

  switch (status) {
    case "DEPLOYMENT_STATUS_PENDING":
      projectStatusStore.set({
        ...DefaultStatusValues,
        pending: true,
        prevStatus: status,
      });
      break;

    case "DEPLOYMENT_STATUS_RECONCILING":
      projectStatusStore.set({
        ...DefaultStatusValues,
        reconciling: true,
        prevStatus: status,
      });
      break;

    case "DEPLOYMENT_STATUS_ERROR":
      projectStatusStore.set({
        ...DefaultStatusValues,
        errored: true,
        prevStatus: status,
      });
      break;

    case "DEPLOYMENT_STATUS_OK":
      projectStatusStore.set({
        ...DefaultStatusValues,
        ok: true,
        prevStatus: status,
      });
      break;
  }
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
  okPollTime = ProjectOkPollTime
) {
  const key = `${orgName}__${projectName}`;
  if (stores.has(key)) return stores.get(key);

  const store = createProjectStatusStore(
    queryClient,
    getProjectQuery,
    reconcilingPollTime,
    erroredPollTime,
    okPollTime
  );
  stores.set(key, store);
  return store;
}

export function createProjectStatusStore(
  queryClient: QueryClient,
  getProjectQuery: CreateQueryResult<V1GetProjectResponse, RpcStatus> & {
    queryKey: QueryKey;
  },
  reconcilingPollTime = ProjectReconcilingPollTime,
  erroredPollTime = ProjectErroredPollTime,
  okPollTime = ProjectOkPollTime
): ProjectStatusStore {
  const projectStatusStore = writable<ProjectStatusState>({
    ...DefaultStatusValues,
  });

  getProjectQuery.subscribe((projectStatusResponse) => {
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
    } else if (state.ok) {
      pollTime = okPollTime;
    }

    if (pollTime) {
      queryClient.setQueryDefaults(getProjectQuery.queryKey, {
        refetchInterval: pollTime,
      });
    }
  });

  return projectStatusStore;
}
