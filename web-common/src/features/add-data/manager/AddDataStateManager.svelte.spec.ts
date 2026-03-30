import { describe, expect, it, beforeEach, vi } from "vitest";
import {
  AddDataStateManager,
  type TransitionEvent,
  TransitionEventType,
} from "@rilldata/web-common/features/add-data/manager/AddDataStateManager.svelte.ts";
import {
  type AddDataState,
  AddDataStep,
  ImportDataStep,
  type ImportStepConfig,
} from "@rilldata/web-common/features/add-data/manager/steps/types.ts";
import { getConnectorDriverForSchema } from "@rilldata/web-common/features/add-data/manager/steps/utils.ts";
import { connectorFormCache } from "@rilldata/web-common/features/add-data/manager/steps/connector.ts";

const ClickhouseSchema = "clickhouse";
const ClickhouseConnector = "clickhouse_conn";
const ClickhouseDriver = getConnectorDriverForSchema(ClickhouseSchema)!;
const ClickhouseImportConfig: ImportStepConfig = {
  importSteps: [
    ImportDataStep.CreateMetricsView,
    ImportDataStep.CreateDashboard,
  ],
  connector: ClickhouseConnector,
  importFrom: {
    from: "table",
    table: "AdBids",
    schema: "default",
    database: "public",
  },
  importTo: {},
  envBlob: null,
};

describe("AddDataStateManager", () => {
  const TestCases: {
    title: string;
    events: {
      event: TransitionEvent;
      expectedStep: AddDataState;
    }[];
  }[] = [
    {
      title:
        "Init => SelectConnector => CreateConnector => ExploreConnector => Import => Done",
      events: [
        {
          event: { type: TransitionEventType.Init },
          expectedStep: { step: AddDataStep.SelectConnector },
        },
        {
          event: {
            type: TransitionEventType.SchemaSelected,
            schema: ClickhouseSchema,
            driver: ClickhouseDriver,
          },
          expectedStep: {
            step: AddDataStep.CreateConnector,
            schema: ClickhouseSchema,
            connectorId: "1",
          },
        },
        {
          event: {
            type: TransitionEventType.ConnectorSelected,
            schema: ClickhouseSchema,
            connector: ClickhouseConnector,
            driver: ClickhouseDriver,
            connectorFormValues: {},
          },
          expectedStep: {
            step: AddDataStep.ExploreConnector,
            schema: ClickhouseSchema,
            connector: ClickhouseConnector,
          },
        },
        {
          event: {
            type: TransitionEventType.ImportConfigured,
            config: ClickhouseImportConfig,
          },
          expectedStep: {
            step: AddDataStep.Import,
            importStep: ImportDataStep.Init,
            config: ClickhouseImportConfig,
          },
        },
        {
          event: { type: TransitionEventType.Imported },
          expectedStep: { step: AddDataStep.Done },
        },
      ],
    },

    {
      title:
        "Init with schema => CreateConnector => ExploreConnector => Import => Done",
      events: [
        {
          event: {
            type: TransitionEventType.Init,
            schema: ClickhouseSchema,
            driver: ClickhouseDriver,
          },
          expectedStep: {
            step: AddDataStep.CreateConnector,
            schema: ClickhouseSchema,
            connectorId: "1",
          },
        },
        {
          event: {
            type: TransitionEventType.ConnectorSelected,
            schema: ClickhouseSchema,
            connector: ClickhouseConnector,
            driver: ClickhouseDriver,
            connectorFormValues: {},
          },
          expectedStep: {
            step: AddDataStep.ExploreConnector,
            schema: ClickhouseSchema,
            connector: ClickhouseConnector,
          },
        },
        {
          event: {
            type: TransitionEventType.ImportConfigured,
            config: ClickhouseImportConfig,
          },
          expectedStep: {
            step: AddDataStep.Import,
            importStep: ImportDataStep.Init,
            config: ClickhouseImportConfig,
          },
        },
        {
          event: { type: TransitionEventType.Imported },
          expectedStep: { step: AddDataStep.Done },
        },
      ],
    },

    {
      title: "Init with connector => ExploreConnector => Import => Done",
      events: [
        {
          event: {
            type: TransitionEventType.Init,
            schema: ClickhouseSchema,
            connector: ClickhouseConnector,
            driver: ClickhouseDriver,
          },
          expectedStep: {
            step: AddDataStep.ExploreConnector,
            schema: ClickhouseSchema,
            connector: ClickhouseConnector,
          },
        },
        {
          event: {
            type: TransitionEventType.ImportConfigured,
            config: ClickhouseImportConfig,
          },
          expectedStep: {
            step: AddDataStep.Import,
            importStep: ImportDataStep.Init,
            config: ClickhouseImportConfig,
          },
        },
        {
          event: { type: TransitionEventType.Imported },
          expectedStep: { step: AddDataStep.Done },
        },
      ],
    },
  ];

  beforeEach(() => {
    connectorFormCache.clear();
  });

  describe("Forward transition", () => {
    TestCases.forEach(({ title, events }) => {
      it(`Test forward transition ${title}`, () => {
        const doneStub = vi.fn();
        const closeStub = vi.fn();
        const stateManager = new AddDataStateManager(
          doneStub,
          closeStub,
          undefined,
        );

        events.forEach(({ event, expectedStep }) => {
          expect(doneStub).not.toHaveBeenCalled();
          stateManager.transition(event);
          expect(stateManager.state).toEqual(expectedStep);
        });

        expect(doneStub).toHaveBeenCalledTimes(1);
        expect(closeStub).not.toHaveBeenCalled();
      });
    });
  });

  describe("Backwards transition", () => {
    TestCases.forEach(({ title, events }) => {
      it(`Test back ${title}`, () => {
        const doneStub = vi.fn();
        const closeStub = vi.fn();
        const stateManager = new AddDataStateManager(
          doneStub,
          closeStub,
          undefined,
        );

        for (let i = 0; i < events.length - 1; i++) {
          stateManager.transition(events[i].event);
        }

        for (let i = events.length - 2; i >= 1; i--) {
          const { expectedStep } = events[i - 1];
          stateManager.transition({ type: TransitionEventType.Back });
          expect(stateManager.state).toEqual(expectedStep);
        }
        stateManager.transition({ type: TransitionEventType.Back });
        expect(stateManager.state.step).toEqual(AddDataStep.Init);

        expect(doneStub).not.toHaveBeenCalled();
        expect(closeStub).toHaveBeenCalledTimes(1);
      });
    });
  });

  it("Lateral transition", () => {
    const doneStub = vi.fn();
    const closeStub = vi.fn();
    const stateManager = new AddDataStateManager(
      doneStub,
      closeStub,
      undefined,
    );

    stateManager.transition({ type: TransitionEventType.Init });
    stateManager.transition({
      type: TransitionEventType.SchemaSelected,
      schema: ClickhouseSchema,
      driver: ClickhouseDriver,
    });
    stateManager.transition({
      type: TransitionEventType.ConnectorSelected,
      schema: ClickhouseSchema,
      connector: ClickhouseConnector,
      driver: ClickhouseDriver,
      connectorFormValues: {},
    });

    // Lateral transition to another connector
    stateManager.transition({
      type: TransitionEventType.ConnectorSelected,
      schema: ClickhouseSchema,
      connector: ClickhouseConnector + "_1",
      driver: ClickhouseDriver,
      connectorFormValues: {},
    });
    // On the correct step
    expect(stateManager.state.step).toEqual(AddDataStep.ExploreConnector);
    expect((stateManager.state as any).connector).toEqual(
      ClickhouseConnector + "_1",
    );

    // Going back will now close the state manager
    stateManager.transition({ type: TransitionEventType.Back });
    expect(stateManager.state.step).toEqual(AddDataStep.Init);
  });
});
