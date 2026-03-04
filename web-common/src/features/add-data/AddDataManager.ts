import { pushState, replaceState } from "$app/navigation";
import type { V1ConnectorDriver } from "@rilldata/web-common/runtime-client";
import {
  connectors,
  getBackendConnectorName,
  getConnectorSchema,
  toConnectorDriver,
} from "@rilldata/web-common/features/sources/modal/connector-schemas.ts";
import { fetchAnalyzeConnectors } from "@rilldata/web-common/features/connectors/selectors.ts";
import { writable } from "svelte/store";

export enum AddDataStep {
  Select,
  Connector,
  Source,
  Explorer,
  Import,
}

export class AddDataManager {
  public stepStore = writable<AddDataStep>(AddDataStep.Select);
  public connectorDriverStore = writable<V1ConnectorDriver | null>(null);
  public schemaNameStore = writable<string | null>(null);
  public connectorNameStore = writable<string | null>(null);

  public constructor(
    private readonly instanceId: string,
    schemaName: string | null,
    connectorName: string | null,
  ) {
    if (connectorName && schemaName) {
      this.selectConnector(schemaName, connectorName, true);
    } else if (schemaName) {
      this.selectSchemaName(schemaName, true);
    } else {
      this.stepStore.set(AddDataStep.Select);
    }
  }

  public selectSchemaName(schemaName: string, shouldReplaceState = false) {
    const connectorDriver = connectorDriverForSchema(schemaName);
    if (!connectorDriver) return;

    const shouldGoToConnectorStep = isConnectorType(connectorDriver);
    const step = shouldGoToConnectorStep
      ? AddDataStep.Connector
      : AddDataStep.Source;

    const state = { step, schema: schemaName };
    if (shouldReplaceState) replaceState("", state);
    else pushState("", state);
  }

  public selectConnector(
    schemaName: string,
    connectorName: string,
    shouldReplaceState = false,
  ) {
    const connectorDriver = connectorDriverForConnector(
      this.instanceId,
      connectorName,
    );
    if (!connectorDriver) return;

    const shouldGoToExplorerStep = isExplorerType(connectorDriver);
    const step = shouldGoToExplorerStep
      ? AddDataStep.Explorer
      : AddDataStep.Source;

    const state = { step, schema: schemaName, connector: connectorName };
    if (shouldReplaceState) replaceState("", state);
    else pushState("", state);
  }

  public applyState(state: any) {
    if (typeof state.step !== "number") return;
    const step = state.step as AddDataStep;
    const schemaName = state.schema as string | undefined;
    const connectorName = state.connector as string | undefined;

    if (connectorName && schemaName) {
      this.setConnectorStep(step, schemaName, connectorName);
    } else if (schemaName) {
      this.setSchemaStep(step, schemaName);
    } else {
      this.stepStore.set(step);
      this.schemaNameStore.set(null);
      this.connectorNameStore.set(null);
    }
  }

  private setSchemaStep(step: AddDataStep, schemaName: string) {
    const connectorDriver = connectorDriverForSchema(schemaName);
    if (!connectorDriver) return;

    this.connectorDriverStore.set(connectorDriver);
    this.schemaNameStore.set(schemaName);
    this.connectorNameStore.set(null);
    this.stepStore.set(step);
  }

  private setConnectorStep(
    step: AddDataStep,
    schemaName: string,
    connectorName: string,
  ) {
    const connectorDriver = connectorDriverForConnector(
      this.instanceId,
      connectorName,
    );
    if (!connectorDriver) return;

    this.connectorDriverStore.set(connectorDriver);
    this.schemaNameStore.set(schemaName);
    this.connectorNameStore.set(connectorName);
    this.stepStore.set(step);
  }
}

export function getPageStateForAddData(schemaName: string | null) {
  if (!schemaName) return { step: AddDataStep.Select };

  const connectorInfo = connectors.find((c) => c.name === schemaName);
  const connectorDriver = connectorInfo
    ? toConnectorDriver(connectorInfo)
    : null;
  if (!connectorDriver) return { step: AddDataStep.Select };

  const shouldGoToConnectorStep = isConnectorType(connectorDriver);
  return {
    step: shouldGoToConnectorStep ? AddDataStep.Connector : AddDataStep.Source,
    schema: schemaName,
  };
}

function connectorDriverForSchema(schemaName: string) {
  const connectorInfo = connectors.find((c) => c.name === schemaName);
  if (!connectorInfo) return null;
  const schema = getConnectorSchema(connectorInfo.name);
  const category = schema?.["x-category"];
  const backendName = getBackendConnectorName(connectorInfo.name);

  return {
    name: backendName,
    displayName: connectorInfo.displayName,
    implementsObjectStore: category === "objectStore",
    implementsOlap: category === "olap",
    implementsSqlStore: category === "sqlStore",
    implementsWarehouse: category === "warehouse",
    implementsFileStore: category === "fileStore",
  };
}

function connectorDriverForConnector(
  instanceId: string,
  connectorName: string,
) {
  const runtimeConnectors = fetchAnalyzeConnectors(instanceId);
  const connectorDriver = runtimeConnectors.find(
    (r) => r.name === connectorName,
  )?.driver;
  return connectorDriver;
}

function isConnectorType(connectorDriver: V1ConnectorDriver) {
  return (
    connectorDriver?.implementsObjectStore ||
    connectorDriver?.implementsOlap ||
    connectorDriver?.implementsSqlStore ||
    (connectorDriver?.implementsWarehouse &&
      connectorDriver?.name !== "salesforce")
    // TODO: multi step?
  );
}

function isExplorerType(connectorDriver: V1ConnectorDriver) {
  return connectorDriver?.implementsOlap || connectorDriver?.implementsSqlStore;
}
