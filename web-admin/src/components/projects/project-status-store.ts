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
};
const DefaultStatusValues = {
  queryRunning: false,
  pending: false,
  reconciling: false,
  errored: false,
  ok: false,
};

export const ProjectReconcilingPollTime = 1000; // 1 sec
export const ProjectErroredPollTime = 5000; // 5 sec
export const ProjectOkPollTime = 60 * 1000; // 1 min

export type ProjectStatusStore = Readable<ProjectStatusState>;

function setValueFromStatus(
  { set }: Writable<ProjectStatusState>,
  status: V1DeploymentStatus
) {
  if (!status) {
    set({
      ...DefaultStatusValues,
      queryRunning: true,
    });
    return;
  }

  switch (status) {
    case "DEPLOYMENT_STATUS_PENDING":
      set({
        ...DefaultStatusValues,
        pending: true,
      });
      break;

    case "DEPLOYMENT_STATUS_RECONCILING":
      set({
        ...DefaultStatusValues,
        reconciling: true,
      });
      break;

    case "DEPLOYMENT_STATUS_ERROR":
      set({
        ...DefaultStatusValues,
        errored: true,
      });
      break;

    case "DEPLOYMENT_STATUS_OK":
      set({
        ...DefaultStatusValues,
        ok: true,
      });
      break;
  }
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
