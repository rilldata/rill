import { beforeAll, beforeEach, describe, it } from "@jest/globals";
import {
  runtimeServiceGetCardinalityOfColumn,
  runtimeServiceGetNumericHistogram,
} from "@rilldata/web-common/runtime-client";
import { httpRequestQueue } from "@rilldata/web-common/runtime-client/http-client";
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

  beforeEach(() => {
    respLock.clear();
    fetchMock.mockClear();
  });

  it("happy path", async () => {
    const table = "t";
    const cols = 5;

    const promises = getProfilingQueries(table, cols);

    await asyncWait(55);
    // only 5 calls go through
    expect(fetchMock.mock.calls.length).toBe(5);
    unlockRequests(table, cols, 0, 2);
    await asyncWait(55);
    expect(fetchMock.mock.calls.length).toBe(7);
    unlockRequests(table, cols, 2);
    await Promise.all(promises);
    expect(fetchMock.mock.calls.length).toBe(10);

    expect(getActualUrls(fetchMock)).toEqual(getProfilingRequests(table, cols));
  });

  it("cancelling queries", async () => {
    const table1 = "t1";
    const cols1 = 5;
    const table2 = "t2";
    const cols2 = 4;

    getProfilingQueries(table1, cols1);

    await asyncWait(55);
    unlockRequests(table1, cols1, 0, 2);
    await asyncWait(55);
    httpRequestQueue.removeByName(table1);
    await asyncWait(55);
    unlockRequests(table1, cols1, 2);
    await asyncWait(55);

    const promises = getProfilingQueries(table2, cols2);
    await asyncWait(55);
    unlockRequests(table2, cols2);

    try {
      await Promise.all(promises);
    } catch (err) {
      // no-op
    }

    expect(getActualUrls(fetchMock)).toEqual([
      ...getProfilingRequests(table1, cols1, 0, 7),
      ...getProfilingRequests(table2, cols2),
    ]);
  });

  it("change priority", async () => {
    const table1 = "t1";
    const cols1 = 5;
    const table2 = "t2";
    const cols2 = 4;

    const promises = getProfilingQueries(table1, cols1);
    await asyncWait(55);
    unlockRequests(table1, cols1, 0, 2);
    await asyncWait(55);
    httpRequestQueue.inactiveByName(table1);
    await asyncWait(55);

    promises.push(...getProfilingQueries(table2, cols2));
    await asyncWait(55);
    unlockRequests(table1, cols1, 2);
    unlockRequests(table2, cols2);
    await Promise.all(promises);

    expect(getActualUrls(fetchMock)).toEqual([
      ...getProfilingRequests(table1, cols1, 0, 7),
      ...getProfilingRequests(table2, cols2),
      ...getProfilingRequests(table1, cols1, 7),
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

  clear() {
    this.lock = new Set<string>();
  },
};

async function mockedQuery(url: string, _entry: RequestQueueEntry) {
  const u = new URL(`http://localhost/${url}`);
  const [, , , , , type, ...parts] = u.pathname.split("/");
  let key: string;

  switch (type) {
    case "queries":
      key = type + "__" + parts[0] + "__" + parts[2] + "__" + (u.searchParams.get("columnName") ?? "");
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

function getActualUrls(fetchMock: Mock) {
  return fetchMock.mock.calls.map((args) =>
    args[0].replace("/v1/instances/i/", "").replace(/&?priority=[0-9]+$/, "")
  );
}

function getProfilingQueries(table: string, cols: number) {
  return [
    ...Array(cols)
      .fill(0)
      .map((_, i) =>
        runtimeServiceGetNumericHistogram("i", table, { columnName: `c${i}` })
      ),
    ...Array(cols)
      .fill(0)
      .map((_, i) =>
        runtimeServiceGetCardinalityOfColumn("i", table, {
          columnName: `c${i}`,
        })
      ),
  ];
}

function getProfilingRequests(
  table: string,
  cols: number,
  start = 0,
  end = -1
) {
  const requests = [
    ...Array(cols)
      .fill(0)
      .map(
        (_, i) => `queries/column-cardinality/tables/${table}?columnName=c${i}`
      ),
    ...Array(cols)
      .fill(0)
      .map(
        (_, i) => `queries/numeric-histogram/tables/${table}?columnName=c${i}`
      ),
  ];
  if (end === -1) {
    return requests.slice(start);
  } else {
    return requests.slice(start, end);
  }
}

function unlockRequests(table: string, cols: number, start = 0, end = -1) {
  const keys = [
    ...Array(cols)
      .fill(0)
      .map((_, i) => `queries__column-cardinality__${table}__c${i}`),
    ...Array(cols)
      .fill(0)
      .map((_, i) => `queries__numeric-histogram__${table}__c${i}`),
  ];
  const endIndex = end === -1 ? keys.length : end;
  for (let i = start; i < endIndex && i < keys.length; i++) {
    respLock.resp(keys[i]);
  }
}
