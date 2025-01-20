import { page } from "$app/stores";
import { derived, get, writable, type Writable } from "svelte/store";
import { queryClient } from "../../../lib/svelte-query/globalQueryClient";
import {
  getRuntimeServiceGetFileQueryKey,
  runtimeServiceDeleteFile,
  runtimeServiceGetFile,
  runtimeServicePutFile,
  runtimeServiceUnpackEmpty,
} from "../../../runtime-client";
import { runtime } from "../../../runtime-client/runtime-store";
import type { OlapDriver } from "../../connectors/olap/olap-config";
import { EMPTY_PROJECT_TITLE } from "../constants";

export class OnboardingState {
  managementType: Writable<"rill-managed" | "self-managed">;
  olapDriver: Writable<OlapDriver>;
  firstDataSource: Writable<string | undefined>;
  runtimeInstanceId: string;

  constructor() {
    this.managementType = writable("rill-managed");
    this.olapDriver = writable("duckdb");
    this.firstDataSource = writable(undefined);
    this.runtimeInstanceId = get(runtime).instanceId;
  }

  async fetch() {
    const response = await queryClient.fetchQuery({
      queryKey: getRuntimeServiceGetFileQueryKey(this.runtimeInstanceId, {
        path: "tmp/onboarding-state.json",
      }),
      queryFn: () =>
        runtimeServiceGetFile(this.runtimeInstanceId, {
          path: "tmp/onboarding-state.json",
        }),
    });

    // if the file doesn't exist, set the state to default
    if (!response.blob) {
      this.managementType.set("rill-managed");
      this.olapDriver.set("duckdb");
      this.firstDataSource.set(undefined);
      return;
    }

    // parse the state
    const state = JSON.parse(response.blob);

    // set the state
    this.managementType.set(state.managementType);
    this.olapDriver.set(state.olapDriver);
    this.firstDataSource.set(state.firstDataSource);
  }

  async save() {
    const state = {
      managementType: get(this.managementType),
      olapDriver: get(this.olapDriver),
      firstDataSource: get(this.firstDataSource),
    };

    const jsonState = JSON.stringify(state);

    await runtimeServicePutFile(this.runtimeInstanceId, {
      path: "tmp/onboarding-state.json",
      blob: jsonState,
    });
  }

  async complete() {
    await runtimeServiceDeleteFile(this.runtimeInstanceId, {
      path: "tmp/onboarding-state.json",
    });
  }

  selectManagementType(type: "rill-managed" | "self-managed") {
    this.managementType.set(type);

    if (type === "rill-managed") {
      this.selectOLAP("duckdb");
    } else {
      this.selectOLAP("clickhouse");
    }

    // Reset the first data source
    this.firstDataSource.set(undefined);

    void this.save();
  }

  selectOLAP(olap: OlapDriver) {
    this.olapDriver.set(olap);

    // reset the first data source
    this.firstDataSource.set(undefined);

    void this.save();
  }

  toggleFirstDataSource(dataSource: string) {
    if (get(this.firstDataSource) === dataSource) {
      this.firstDataSource.set(undefined);
    } else {
      this.firstDataSource.set(dataSource);
    }

    void this.save();
  }

  async skipFirstSource() {
    // unpack the empty project
    await runtimeServiceUnpackEmpty(this.runtimeInstanceId, {
      displayName: EMPTY_PROJECT_TITLE,
    });

    // create an OLAP connector file
    await runtimeServicePutFile(this.runtimeInstanceId, {
      path: "rill.yaml",
      blob: `type: connector
driver: ${get(this.olapDriver)}`,
    });

    // Edit the rill.yaml file
    // TODO
  }

  // Clean up all files created by the Add Data form
  async cleanUp() {
    await runtimeServiceDeleteFile(this.runtimeInstanceId, {
      path: "connectors",
      force: true,
    });

    await runtimeServiceDeleteFile(this.runtimeInstanceId, {
      path: "rill.yaml",
    });

    await runtimeServiceDeleteFile(this.runtimeInstanceId, {
      path: ".env",
    });

    await runtimeServiceDeleteFile(this.runtimeInstanceId, {
      path: ".gitignore",
    });
  }

  // TODO: Implement this
  isInitialized() {
    return false;
  }

  getNumberOfSteps = () => {
    return derived(this.managementType, (managementType) =>
      managementType === "rill-managed" ? 2 : 3,
    );
  };

  getCurrentStep = () => {
    // /welcome/select-connectors => 1
    // /welcome/add-credentials => 2
    // /welcome/make-your-first-dashboard => 3
    return derived(page, (page) => {
      if (page.url.pathname.includes("select-connectors")) {
        return 1;
      }
      if (page.url.pathname.includes("add-credentials")) {
        return 2;
      }
      if (page.url.pathname.includes("make-your-first-dashboard")) {
        return 3;
      }

      return 0;
    });
  };
}

export function getOnboardingState() {
  return new OnboardingState();
}
