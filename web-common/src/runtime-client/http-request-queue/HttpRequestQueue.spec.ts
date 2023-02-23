import { beforeAll, beforeEach, describe, it } from "@jest/globals";
import {
  runtimeServiceGetCardinalityOfColumn,
  runtimeServiceGetNumericHistogram,
  RuntimeServiceGetNumericHistogramHistogramMethod,
} from "@rilldata/web-common/runtime-client";
import { httpRequestQueue } from "@rilldata/web-common/runtime-client/http-client";
import { UrlExtractorRegex } from "@rilldata/web-common/runtime-client/http-request-queue/HttpRequestQueue";
import type { RequestQueueEntry } from "@rilldata/web-common/runtime-client/http-request-queue/HttpRequestQueueTypes";
import { asyncWait, waitUntil } from "@rilldata/web-local/lib/util/waitUtils";
import Mock = jest.Mock;

// skipping because there is too much instability due to race conditions
// TODO: figure out a good way to test
describe.skip("HttpRequestQueue", () => {
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
    expect(fetchMock.mock.calls.length).toBe(cols * 2);

    expect(correctActualUrls(fetchMock)).toEqual(
      getProfilingRequests(table, cols)
    );
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

    expect(correctActualUrls(fetchMock)).toEqual([
      ...getProfilingRequests(table1, cols1, 0, 5),
      ...getProfilingRequests(table1, cols1, 6, 8),
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

    expect(correctActualUrls(fetchMock)).toEqual([
      ...getProfilingRequests(table1, cols1, 0, 5),
      ...getProfilingRequests(table1, cols1, 6, 8),
      ...getProfilingRequests(table2, cols2),
      ...getProfilingRequests(table1, cols1, 5, 6),
      ...getProfilingRequests(table1, cols1, 8),
    ]);
  });

  it("change column priority", async () => {
    const table = "t";
    const cols = 8;

    const promises = getProfilingQueries(table, cols);

    await asyncWait(55);
    // only 5 calls go through
    expect(fetchMock.mock.calls.length).toBe(5);
    unlockRequests(table, cols, 0, 2);
    await asyncWait(55);
    httpRequestQueue.prioritiseColumn(table, "c2", true);
    unlockRequests(table, cols, 2);
    await Promise.all(promises);
    expect(fetchMock.mock.calls.length).toBe(cols * 2);

    expect(correctActualUrls(fetchMock)).toEqual([
      ...getProfilingRequests(table, cols, 0, 5),
      ...getProfilingRequests(table, cols, 6, 8),
      ...getProfilingRequests(table, cols, 10, 11),
      ...getProfilingRequests(table, cols, 5, 6),
      ...getProfilingRequests(table, cols, 8, 10),
      ...getProfilingRequests(table, cols, 11),
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
      key =
        type +
        "__" +
        parts[0] +
        "__" +
        parts[2] +
        "__" +
        (u.searchParams.get("columnName") ?? "");
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

/**
 * Requests within a query type is not always in order.
 * Correct it by sorting individual types.
 */
function correctActualUrls(fetchMock: Mock): Array<string> {
  const actualUrls = fetchMock.mock.calls.map((args) => args[0]);
  const groups = new Array<Array<string>>();

  let lastName: string;
  let lastType: string;
  let lastGroup: Array<string>;
  actualUrls.forEach((actualUrl) => {
    actualUrl = actualUrl.replace(/&?priority=[0-9]+$/, "");
    const urlMatch = UrlExtractorRegex.exec(actualUrl.replace(/\?.*$/, ""));
    let name: string;
    let type: string;
    switch (urlMatch?.[1]) {
      case "metrics-views":
        name = urlMatch[3];
        type = urlMatch[2];
        break;
      case "queries":
        name = urlMatch[4];
        type = urlMatch[2];
    }
    if (lastType !== type || lastName !== name) {
      if (lastGroup?.length) {
        lastGroup.sort();
        groups.push(lastGroup);
      }
      lastName = name;
      lastType = type;
      lastGroup = [];
    }
    lastGroup.push(actualUrl.replace("/v1/instances/i/", ""));
  });
  if (lastGroup?.length) {
    lastGroup.sort();
    groups.push(lastGroup);
  }

  return groups.flat();
}

function getProfilingQueries(table: string, cols: number) {
  return [
    ...Array(cols)
      .fill(0)
      .map((_, i) =>
        runtimeServiceGetNumericHistogram("i", table, {
          columnName: `c${i}`,
          histogramMethod:
            RuntimeServiceGetNumericHistogramHistogramMethod.HISTOGRAM_METHOD_FD,
        })
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
  startInclusive = 0,
  endExclusive = -1
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
  if (endExclusive === -1) {
    return requests.slice(startInclusive);
  } else {
    return requests.slice(startInclusive, endExclusive);
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
