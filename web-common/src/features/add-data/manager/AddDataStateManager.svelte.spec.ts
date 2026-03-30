import { describe, expect, it } from "vitest";
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
      expectedStep?: AddDataState;
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
            assignedConnectorName: "clickhouse",
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
            assignedConnectorName: "clickhouse",
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

  describe("Forward transition", () => {
    TestCases.forEach(({ title, events }) => {
      it(`Test forward transition ${title}`, () => {
        const stateManager = new AddDataStateManager(
          undefined,
          undefined,
          undefined,
        );

        events.forEach(({ event, expectedStep }) => {
          const prevState = { ...stateManager.state };
          stateManager.transition(event);
          if (expectedStep) expect(stateManager.state).toEqual(expectedStep);
          else expect(stateManager.state).toEqual(prevState);
        });
      });
    });
  });

  describe("Backwards transition", () => {
    TestCases.forEach(({ title, events }) => {
      it(`Test back ${title}`, () => {
        const stateManager = new AddDataStateManager(
          undefined,
          undefined,
          undefined,
        );

        for (let i = 0; i < events.length - 1; i++) {
          stateManager.transition(events[i].event);
        }

        for (let i = events.length - 3; i >= 0; i--) {
          const { expectedStep } = events[i];
          stateManager.transition({ type: TransitionEventType.Back });
          if (expectedStep) expect(stateManager.state).toEqual(expectedStep);
        }
      });
    });
  });
});
