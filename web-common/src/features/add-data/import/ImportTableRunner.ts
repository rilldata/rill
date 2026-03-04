import {
  runtimeServicePutFileAndWaitForReconciliation,
  waitForResourceReconciliation,
} from "@rilldata/web-common/features/entity-management/actions.ts";
import { getFileAPIPathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers.ts";
import { EntityType } from "@rilldata/web-common/features/entity-management/types.ts";
import {
  runtimeServiceGenerateMetricsViewFile,
  runtimeServicePutFile,
} from "@rilldata/web-common/runtime-client";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
import type { ConnectorTableEntry } from "@rilldata/web-common/features/connectors/explorer/connector-explorer-store.ts";
import { get, writable } from "svelte/store";
import { getName } from "@rilldata/web-common/features/entity-management/name-utils.ts";
import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts.ts";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
import { createResourceFile } from "@rilldata/web-common/features/file-explorer/new-files.ts";
import { waitUntil } from "@rilldata/web-common/lib/waitUtils.ts";
import { goto } from "$app/navigation";
import { splitFolderFileNameAndExtension } from "@rilldata/web-common/features/entity-management/file-path-utils.ts";

export enum ImportTableMode {
  Init,
  CreateModel,
  CreateMetrics,
  CreateExplore,
}

export class ImportTableRunner {
  public mode = writable<ImportTableMode>(ImportTableMode.Init);
  public error = writable<string | null>(null);
  public details = writable<string | null>(null);
  public currentFilePath = writable<string | null>(null);

  public constructor(
    private readonly instanceId: string,
    private readonly name: string,
    private readonly connectorTableEntry: ConnectorTableEntry,
    private readonly yaml: string,
    private readonly envBlob: string | null,
  ) {}

  public async run() {
    try {
      this.mode.set(ImportTableMode.Init);
      const filePath = getFileAPIPathFromNameAndType(
        this.name,
        EntityType.Model,
        true,
      );
      this.currentFilePath.set(filePath);

      this.mode.set(ImportTableMode.CreateModel);
      await runtimeServicePutFile(this.instanceId, {
        path: filePath,
        blob: this.yaml,
        create: true,
        createOnly: false,
      });

      if (this.envBlob !== null) {
        // Make sure the file has reconciled before testing the connection
        await runtimeServicePutFileAndWaitForReconciliation(this.instanceId, {
          path: ".env",
          blob: this.envBlob,
          create: true,
          createOnly: false,
        });
      }

      // Wait for the model to successfully reconcile
      await waitForResourceReconciliation(
        this.instanceId,
        this.name,
        ResourceKind.Model,
      );

      // Metrics view generation
      this.mode.set(ImportTableMode.CreateMetrics);
      const newMetricsViewName = getName(
        `${this.name}_metrics`,
        fileArtifacts.getNamesForKind(ResourceKind.MetricsView),
      );
      const newMetricsViewFilePath = `/metrics/${newMetricsViewName}.yaml`;
      this.currentFilePath.set(newMetricsViewFilePath);

      // Call GenerateMetricsViewFile with the generated file path
      await runtimeServiceGenerateMetricsViewFile(this.instanceId, {
        table: this.name,
        connector: this.connectorTableEntry.connector,
        database: this.connectorTableEntry.database,
        databaseSchema: this.connectorTableEntry.schema,
        path: newMetricsViewFilePath,
        useAi: false, // TODO: check feature flags
      });
      // Wait for the metrics view to successfully reconcile
      await waitForResourceReconciliation(
        this.instanceId,
        newMetricsViewName,
        ResourceKind.MetricsView,
      );

      // Explore generation
      this.mode.set(ImportTableMode.CreateExplore);
      // Get the MetricsView resource used to create the explore from.
      const metricsViewResourceResp = fileArtifacts
        .getFileArtifact(newMetricsViewFilePath)
        .getResource(queryClient, this.instanceId);
      await waitUntil(
        () => get(metricsViewResourceResp).data !== undefined,
        5000,
      );
      const metricsViewResource = get(metricsViewResourceResp).data;
      if (!metricsViewResource) {
        throw new Error("Failed to create a Metrics View resource");
      }

      // Create the Explore file
      const exploreFilePath = await createResourceFile(
        ResourceKind.Explore,
        metricsViewResource,
      );
      this.currentFilePath.set(exploreFilePath);

      // Get the explore name and wait for it to reconcile
      const [, exploreName] = splitFolderFileNameAndExtension(exploreFilePath);
      await waitForResourceReconciliation(
        this.instanceId,
        exploreName,
        ResourceKind.Explore,
      );

      // Go to the explore preview directly
      await goto(`/explore/${exploreName}`);
    } catch (error) {
      this.error.set(error?.response?.data?.message ?? error?.message ?? null);
      this.details.set(error?.details ?? null);
      throw error;
    }
  }
}
