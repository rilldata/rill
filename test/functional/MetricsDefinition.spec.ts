import { TestBase } from "@adityahegde/typescript-test-utils";
import type { TestServerSetupParameter } from "../utils/ServerSetup";
import { JestTestLibrary } from "@adityahegde/typescript-test-utils/dist/jest/JestTestLibrary";
import { TestServerSetup } from "../utils/ServerSetup";
import { readFileSync, writeFileSync } from "fs";
import { TwoTableJoinQuery } from "../data/ModelQuery.data";
import { asyncWait } from "$common/utils/waitUtils";
import axios from "axios";
import type { PersistentModelState } from "$common/data-modeler-state-service/entity-state-service/PersistentModelEntityService";
import { HttpStreamClient } from "$lib/http-client/HttpStreamClient";

const CLI_FOLDER = "temp/metrics-def";
const PORT = 8080;
const URL = `http://localhost:${PORT}`;

@TestBase.ParameterizedSuite([
  {
    cliFolder: CLI_FOLDER,
    serverPort: PORT,
  } as TestServerSetupParameter,
])
@TestBase.TestLibrary(JestTestLibrary)
@TestBase.TestSuiteSetup(TestServerSetup)
export class MetricsDefinitionSpec extends TestBase<TestServerSetupParameter> {
  private httpStreamClient = new HttpStreamClient(`${URL}/api`, () => {});

  @TestBase.BeforeSuite()
  public async setupFiles() {
    writeFileSync(`${CLI_FOLDER}/models/query_1.sql`, TwoTableJoinQuery);
    await asyncWait(200);
  }

  @TestBase.Test()
  public async createMetricsDefinition() {
    const resp = await axios.put(
      `${URL}/api/metrics`,
      {},
      {
        headers: { "Content-Type": "application/json" },
      }
    );
    const id = resp.data.match(/"id":"(.*?)"/)[1];

    const modelState: PersistentModelState = JSON.parse(
      readFileSync(`${CLI_FOLDER}/state/derived_model_state.json`).toString()
    );

    await axios.post(`/metrics/${id}/updateModel`, {
      modelId: modelState.entities[0].id,
    });
  }
}
