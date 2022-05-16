import { DataProviderData, TestBase } from "@adityahegde/typescript-test-utils";
import { JestTestLibrary } from "@adityahegde/typescript-test-utils/dist/jest/JestTestLibrary";
import { QueryParser } from "$common/query-parser/QueryParser";
import {
  CTE,
  NestedQuery,
  SingleTableQuery,
  TwoTableJoinQuery,
} from "../data/ModelQuery.data";
import type { QueryTreeJSON } from "$common/query-parser/tree/QueryTree";
import {
  CTEQueryTree,
  NestedQueryTree,
  SingleTableQueryTree,
  TwoTableJoinQueryTree,
} from "../data/QueryParser.data";

@TestBase.Suite
@TestBase.TestLibrary(JestTestLibrary)
export class QueryParserSpec extends TestBase {
  public queryParseData(): DataProviderData<[string, QueryTreeJSON]> {
    return {
      subData: [
        {
          title: "SingleTableQuery",
          args: [SingleTableQuery, SingleTableQueryTree],
        },
        {
          title: "TwoTableJoinQuery",
          args: [TwoTableJoinQuery, TwoTableJoinQueryTree],
        },
        {
          title: "NestedQuery",
          args: [NestedQuery, NestedQueryTree],
        },
        {
          title: "CTE",
          args: [CTE, CTEQueryTree],
        },
      ],
    };
  }

  @TestBase.Test("queryParseData")
  public shouldParseQuery(
    query: string,
    expectedJson: Record<string, QueryTreeJSON>
  ) {
    const parser = new QueryParser();
    expect(this.stripOffLocation(parser.parse(query).toJSON())).toEqual(
      expectedJson
    );
  }

  // it is not practical to include locations in assertion
  // as any change would need tedious updating of the expected locations
  private stripOffLocation(json) {
    if ("start" in json) delete json.start;
    if ("end" in json) delete json.end;

    for (const key in json) {
      if (typeof json[key] === "object") {
        json[key] = this.stripOffLocation(json[key]);
      }
    }

    return json;
  }
}
