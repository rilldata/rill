import { describe } from "@jest/globals";
import {
  runtimeServiceGetCardinalityOfColumn,
  runtimeServiceGetNumericHistogram,
  runtimeServiceGetTableCardinality,
} from "@rilldata/web-common/runtime-client";
import {
  asyncWait,
  waitUntil,
} from "@rilldata/web-local/common/utils/waitUtils";
import type { RequestQueueEntry } from "@rilldata/web-local/lib/http-request-queue/HttpRequestQueueTypes";
import Mock = jest.Mock;

describe("HttpRequestQueue", () => {
  let fetchMock: Mock;
  let originalFetch;
  beforeAll(() => {
    fetchMock = jest.fn(mockedQuery);
    originalFetch = global.fetch;
    global.fetch = fetchMock;
  });

  it("happy path", async () => {
    const promises = [
      runtimeServiceGetNumericHistogram("i", "t", "c1"),
      runtimeServiceGetNumericHistogram("i", "t", "c2"),
      runtimeServiceGetCardinalityOfColumn("i", "t", "c1"),
      runtimeServiceGetTableCardinality("i", "t"),
      runtimeServiceGetNumericHistogram("i", "t", "c3"),
      runtimeServiceGetCardinalityOfColumn("i", "t", "c2"),
      runtimeServiceGetCardinalityOfColumn("i", "t", "c3"),
    ];

    await asyncWait(100);
    // only 5 calls go through
    expect(fetchMock.mock.calls.length).toBe(5);
    respLock.resp("queries__numeric-histogram__t__c1");
    respLock.resp("queries__numeric-histogram__t__c2");
    await asyncWait(100);
    expect(fetchMock.mock.calls.length).toBe(7);
    respLock.resp("queries__numeric-histogram__t__c3");
    respLock.resp("queries__cardinality__t__");
    respLock.resp("queries__cardinality__t__c1");
    respLock.resp("queries__cardinality__t__c2");
    respLock.resp("queries__cardinality__t__c3");
    await Promise.all(promises);

    expect(
      fetchMock.mock.calls.map((args) =>
        args[0].replace("/v1/instances/i/", "")
      )
    ).toEqual([
      "queries/numeric-histogram/tables/t/columns/c1",
      "queries/numeric-histogram/tables/t/columns/c2",
      "queries/numeric-histogram/tables/t/columns/c3",
      "queries/cardinality/tables/t/columns/c1",
      "queries/cardinality/tables/t/columns/c2",
      "queries/cardinality/tables/t/columns/c3",
      "queries/cardinality/tables/t",
    ]);
  });

  afterAll(() => {
    if (originalFetch) {
      global.fetch = originalFetch;
    }
  });
});

const respLock = {
  lock: new Set<string>(),

  async wait(key): Promise<boolean> {
    return waitUntil(() => this.lock.has(key), 5000, 50);
  },

  resp(key) {
    this.lock.add(key);
  },
};

async function mockedQuery(url: string, _entry: RequestQueueEntry) {
  const u = new URL(`http://localhost/${url}`);
  const [, , , , , type, ...parts] = u.pathname.split("/");
  let key: string;

  switch (type) {
    case "queries":
      key = type + "__" + parts[0] + "__" + parts[2] + "__" + (parts[4] ?? "");
      break;

    case "metrics-views":
      key = type + "__" + parts[0] + "__" + parts[1];
      break;

    default:
      key = url;
      break;
  }

  await respLock.wait(key);

  return {
    ok: true,
    json: () => url,
  };
}
