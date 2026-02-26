import { waitForResourceReconciliation } from "@rilldata/web-common/features/entity-management/actions.ts";
import { getFileAPIPathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers.ts";
import { EntityType } from "@rilldata/web-common/features/entity-management/types.ts";
import { runtimeServicePutFile } from "@rilldata/web-common/runtime-client";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
import type { ConnectorTableEntry } from "@rilldata/web-common/features/connectors/explorer/connector-explorer-store.ts";
import { useCreateMetricsViewFromTableUIAction } from "@rilldata/web-common/features/metrics-views/ai-generation/generateMetricsView.ts";
import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes.ts";
import { MetricsEventSpace } from "@rilldata/web-common/metrics/service/MetricsTypes.ts";
import { writable, type Writable } from "svelte/store";

export enum ImportTableMode {
  Init,
  CreateModel,
  CreateMetrics,
  CreateExplore,
}

export class ImportTableRunner {
  public mode: Writable<ImportTableMode> = writable(ImportTableMode.Init);
  public error: string | null = null;

  public constructor(
    private readonly instanceId: string,
    private readonly name: string,
    private readonly connectorTableEntry: ConnectorTableEntry,
    private readonly yaml: string,
  ) {}

  public async run() {
    try {
      this.mode.set(ImportTableMode.Init);
      const filePath = getFileAPIPathFromNameAndType(
        this.name,
        EntityType.Model,
      );

      await runtimeServicePutFile(this.instanceId, {
        path: filePath,
        blob: this.yaml,
        create: true,
        createOnly: false,
      });
      this.mode.set(ImportTableMode.CreateModel);

      await waitForResourceReconciliation(
        this.instanceId,
        this.name,
        ResourceKind.Model,
      );

      this.mode.set(ImportTableMode.CreateMetrics);

      const creator = useCreateMetricsViewFromTableUIAction(
        this.instanceId,
        this.connectorTableEntry.connector,
        this.connectorTableEntry.database,
        "",
        this.name,
        true,
        BehaviourEventMedium.Button,
        MetricsEventSpace.Workspace, // TODO
        () => {
          this.mode.set(ImportTableMode.CreateExplore);
        },
        false,
      );
      await creator();
    } catch (error) {
      this.error = error?.response?.data?.message ?? error?.message ?? null;
      throw error;
    }
  }
}
