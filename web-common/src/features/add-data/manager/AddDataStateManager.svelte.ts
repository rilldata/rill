import {
  type AddDataState,
  AddDataStep,
  ImportDataStep,
  type ImportStepConfig,
} from "@rilldata/web-common/features/add-data/manager/steps/types.ts";
import {
  isConnectorType,
  isExplorerType,
} from "@rilldata/web-common/features/add-data/manager/steps/utils.ts";
import type { V1ConnectorDriver } from "@rilldata/web-common/runtime-client";
import { getName } from "@rilldata/web-common/features/entity-management/name-utils.ts";
import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts.ts";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";

export enum TransitionEventType {
  Init,
  SchemaSelected,
  // Fired when either connector is created or it is selected to create a model/metrics view on.
  ConnectorSelected,
  ImportConfigured,
  Imported,
  Back,
}
type SchemaSelectedEvent = {
  type: TransitionEventType.SchemaSelected;
  schema: string;
  driver: V1ConnectorDriver;
};
type ConnectorSelectedEvent = {
  type: TransitionEventType.ConnectorSelected;
  connector: string;
  schema: string;
  driver: V1ConnectorDriver;
  connectorFormValues: Record<string, any>;
};
export type TransitionEvent =
  | {
      type: TransitionEventType.Init;
      connector?: string;
      schema?: string;
      driver?: V1ConnectorDriver;
    }
  | SchemaSelectedEvent
  | ConnectorSelectedEvent
  | { type: TransitionEventType.ImportConfigured; config: ImportStepConfig }
  | { type: TransitionEventType.Imported }
  | { type: TransitionEventType.Back };

export class AddDataStateManager {
  public state = $state<AddDataState>({ step: AddDataStep.Init });
  public stateStack = $state<AddDataState[]>([]);

  public constructor(
    private readonly onDone: (() => void) | undefined,
    private readonly onClose: (() => void) | undefined,
    private readonly onStepChange: ((step: AddDataStep) => void) | undefined,
  ) {}

  public transition(event: TransitionEvent) {
    switch (this.state.step) {
      // Init =={Init event}=> SelectConnector
      // Init =={Init event with schema}=> CreateConnector/CreateModel
      // Init =={Init event with connector}=> CreateModel/ExploreConnector
      case AddDataStep.Init: {
        if (event.type !== TransitionEventType.Init) return;
        let newState: AddDataState = { step: AddDataStep.SelectConnector };
        if (event.driver && event.schema && event.connector) {
          newState = getStepForConnector(
            event.schema,
            event.connector,
            event.driver,
          );
        } else if (event.driver && event.schema) {
          newState = getStepForSchema(event.schema, event.driver);
        }
        this.pushState(newState);
        break;
      }

      // SelectConnector =={SchemaSelected event}=> CreateConnector/CreateModel
      case AddDataStep.SelectConnector: {
        if (event.type !== TransitionEventType.SchemaSelected) return;
        const newState = getStepForSchema(event.schema, event.driver);
        this.pushState(newState);
        break;
      }

      // CreateConnector =={Back event}=> Init/SelectConnector
      // CreateConnector =={ConnectorSelected event}=> CreateModel/ExploreConnector
      case AddDataStep.CreateConnector: {
        if (event.type === TransitionEventType.Back) {
          this.popState();
          return;
        }
        if (event.type !== TransitionEventType.ConnectorSelected) return;
        const newState = getStepForConnector(
          event.schema,
          event.connector,
          event.driver,
        );
        this.pushState(newState);
        break;
      }

      // CreateModel/ExploreConnector =={Back event}=> Init/SelectConnector/CreateConnector
      // CreateModel/ExploreConnector =={ImportConfigured event}=> Import
      case AddDataStep.CreateModel:
      case AddDataStep.ExploreConnector:
        if (event.type === TransitionEventType.Back) {
          // Can be back to Init/SelectConnector/CreateConnector
          this.popState();
          return;
        }
        if (event.type !== TransitionEventType.ImportConfigured) return;
        this.pushState({
          step: AddDataStep.Import,
          importStep: ImportDataStep.Init,
          config: event.config,
        });
        break;

      // Import =={Back event}=> Init/CreateModel/ExploreConnector
      // Import =={Imported event}=> Done
      case AddDataStep.Import:
        if (event.type === TransitionEventType.Back) {
          // Can be back to Init/CreateModel/ExploreConnector
          this.popState();
          return;
        }
        if (event.type !== TransitionEventType.Imported) return;
        this.pushState({
          step: AddDataStep.Done,
        });
        break;
    }
  }

  private pushState(state: AddDataState) {
    if (state.step === AddDataStep.Done) this.onDone?.();
    // Only add to the stack if the step changed.
    // Step params change like schema/connector name shouldn't add a new stack entry.
    if (this.state.step !== state.step) this.stateStack.push(this.state);
    this.state = state;
    this.onStepChange?.(state.step);
  }

  private popState() {
    if (this.stateStack.length === 0) this.onClose?.();
    this.state = this.stateStack.pop() ?? { step: AddDataStep.Done };
    this.onStepChange?.(this.state.step);
  }
}

function getStepForSchema(
  schema: string,
  driver: V1ConnectorDriver,
): AddDataState {
  if (isConnectorType(driver)) {
    return {
      step: AddDataStep.CreateConnector,
      schema,
      assignedConnectorName: getConnectorName(schema),
    };
  } else {
    return {
      step: AddDataStep.CreateModel,
      schema,
      connector: driver.name!,
      connectorFormValues: {},
    };
  }
}

function getStepForConnector(
  schema: string,
  connector: string,
  driver: V1ConnectorDriver,
): AddDataState {
  if (isExplorerType(driver))
    return {
      step: AddDataStep.ExploreConnector,
      schema,
      connector,
    };
  else {
    return {
      step: AddDataStep.CreateModel,
      schema,
      connector,
      connectorFormValues: {},
    };
  }
}

function getConnectorName(schema: string) {
  return getName(schema, fileArtifacts.getNamesForKind(ResourceKind.Connector));
}
