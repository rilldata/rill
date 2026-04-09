import {
  type AddDataConfig,
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
import { connectorFormCache } from "@rilldata/web-common/features/add-data/manager/steps/connector.ts";
import {
  behaviourEvent,
  errorEventHandler,
} from "@rilldata/web-common/metrics/initMetrics.ts";
import {
  type AddDataBehaviourEventFields,
  BehaviourEventAction,
} from "@rilldata/web-common/metrics/service/BehaviourEventTypes.ts";

export enum TransitionEventType {
  Init,
  SchemaSelected,
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

  private onDone: (() => void) | undefined = undefined;
  private onClose: (() => void) | undefined = undefined;
  private onStepChange: ((step: AddDataStep) => void) | undefined = undefined;
  private config: AddDataConfig | undefined = undefined;

  private startTime = Date.now();

  public setCallbacks(
    onDone: (() => void) | undefined,
    onClose: (() => void) | undefined,
    onStepChange: ((step: AddDataStep) => void) | undefined,
  ) {
    this.onDone = onDone;
    this.onClose = onClose;
    this.onStepChange = onStepChange;
  }

  public setConfig(config: AddDataConfig) {
    this.config = config;
  }

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
        if (event.type === TransitionEventType.Back) {
          this.popState();
          break;
        }
        if (event.type !== TransitionEventType.SchemaSelected) return;
        const newState = getStepForSchema(event.schema, event.driver);
        this.pushState(newState);
        break;
      }

      // CreateConnector =={Back event}=> Init/SelectConnector
      // CreateConnector =={ConnectorSelected event}=> CreateModel/ExploreConnector
      case AddDataStep.CreateConnector: {
        if (event.type === TransitionEventType.Back) {
          this.fireBehaviourEvent(
            BehaviourEventAction.ConnectorConfigurationCanceled,
            { step: AddDataStep.CreateConnector, schema: this.state.schema },
          );
          this.popState();
          break;
        }
        if (event.type !== TransitionEventType.ConnectorSelected) return;
        const newState = getStepForConnector(
          event.schema,
          event.connector,
          event.driver,
          event.connectorFormValues,
        );
        this.pushState(newState);
        break;
      }

      case AddDataStep.CreateModel:
      case AddDataStep.ExploreConnector:
        switch (event.type) {
          // CreateModel/ExploreConnector =={Back event}=> Init/SelectConnector/CreateConnector
          case TransitionEventType.Back:
            this.fireBehaviourEvent(
              this.state.step === AddDataStep.CreateModel
                ? BehaviourEventAction.ModelConfigurationCanceled
                : BehaviourEventAction.ConnectorExploreCanceled,
              {
                step: this.state.step,
                schema: this.state.schema,
                connector: this.state.connector,
              },
            );
            this.popState();
            break;

          // CreateModel/ExploreConnector =={ImportConfigured event}=> Import
          case TransitionEventType.ImportConfigured:
            this.pushState({
              step: AddDataStep.Import,
              schema: this.state.schema,
              importStep: ImportDataStep.Init,
              config: event.config,
            });
            break;

          // CreateModel/ExploreConnector =={ConnectorSelected event}=> CreateModel/ExploreConnector
          // Connector selected from the connector dropdown in the header.
          case TransitionEventType.ConnectorSelected:
            this.clearStack(); // Lateral state change, clear the stack.
            this.pushState(
              getStepForConnector(
                event.schema,
                event.connector,
                event.driver,
                event.connectorFormValues,
              ),
            );
            break;

          // CreateModel/ExploreConnector =={SchemaSelected event}=> CreateModel/ExploreConnector
          // New connector selected from the connector dropdown in the header.
          case TransitionEventType.SchemaSelected:
            this.clearStack(); // Lateral state change, clear the stack.
            this.pushState(getStepForSchema(event.schema, event.driver));
            break;
        }
        break;

      // Import =={Back event}=> Init/CreateModel/ExploreConnector
      // Import =={Imported event}=> Done
      case AddDataStep.Import:
        if (event.type === TransitionEventType.Back) {
          this.fireBehaviourEvent(BehaviourEventAction.ImportCanceled, {
            step: AddDataStep.Import,
            schema: this.state.schema,
            connector: this.state.config.connector,
          });
          // Can be back to Init/CreateModel/ExploreConnector
          this.popState();
          break;
        }
        if (event.type !== TransitionEventType.Imported) return;
        this.pushState({
          step: AddDataStep.Done,
        });
        break;
    }

    if (event.type !== TransitionEventType.Back) {
      switch (this.state.step) {
        case AddDataStep.CreateConnector:
          this.fireBehaviourEvent(
            BehaviourEventAction.ConnectorConfigurationStarted,
            {
              step: AddDataStep.CreateConnector,
              schema: this.state.schema,
            },
          );
          break;

        case AddDataStep.CreateModel:
          this.fireBehaviourEvent(
            BehaviourEventAction.ModelConfigurationStarted,
            {
              step: AddDataStep.CreateModel,
              schema: this.state.schema,
              connector: this.state.connector,
            },
          );
          break;

        case AddDataStep.ExploreConnector:
          this.fireBehaviourEvent(
            BehaviourEventAction.ConnectorExploreStarted,
            {
              step: AddDataStep.ExploreConnector,
              schema: this.state.schema,
              connector: this.state.connector,
            },
          );
          break;

        case AddDataStep.Import:
          this.fireBehaviourEvent(BehaviourEventAction.ImportStarted, {
            step: AddDataStep.Import,
            schema: this.state.schema,
            connector: this.state.config.connector,
          });
          break;
      }
    }
  }

  public fireErrorEvent(
    message: string,
    step: AddDataStep | ImportDataStep = this.state.step,
  ) {
    if (!this.config?.space || !this.config?.screen) return;

    const addDataFields: AddDataBehaviourEventFields = {
      step,
      duration: Date.now() - this.startTime,
    };
    if (
      this.state.step === AddDataStep.CreateConnector ||
      this.state.step === AddDataStep.CreateModel ||
      this.state.step === AddDataStep.ExploreConnector ||
      this.state.step === AddDataStep.Import
    ) {
      addDataFields.schema = this.state.schema;
    }
    if (
      this.state.step === AddDataStep.CreateModel ||
      this.state.step === AddDataStep.ExploreConnector
    ) {
      addDataFields.connector = this.state.connector;
    }
    if (this.state.step === AddDataStep.Import) {
      addDataFields.connector = this.state.config.connector;
    }

    void errorEventHandler?.fireAddDataErrorEvent(
      this.config.space,
      this.config.screen,
      message,
      addDataFields,
    );
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
    this.state = this.stateStack.pop() ?? { step: AddDataStep.Init };
    if (this.stateStack.length === 0) this.onClose?.();
    this.onStepChange?.(this.state.step);
  }

  // For lateral state change, going back is not supported.
  // So we need to clear the stack.
  private clearStack() {
    this.stateStack = [];
  }

  private fireBehaviourEvent(
    action: BehaviourEventAction,
    fields: AddDataBehaviourEventFields,
  ) {
    if (!this.config?.medium || !this.config?.space || !this.config?.screen)
      return;
    void behaviourEvent?.fireAddDataStepEvent(
      action,
      this.config.medium,
      this.config.space,
      this.config.screen,
      {
        ...fields,
        duration: Date.now() - this.startTime,
      },
    );
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
      connectorId: connectorFormCache.getNextId(),
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
  connectorFormValues: Record<string, any> = {},
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
      connectorFormValues,
    };
  }
}
