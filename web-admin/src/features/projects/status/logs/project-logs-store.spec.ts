import { V1LogLevel } from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";
import { ProjectLogsStore } from "./project-logs-store";

describe("ProjectLogsStore", () => {
  it("addLog assigns a monotonic _id counter", () => {
    const store = new ProjectLogsStore(100);
    const a = store.addLog({ message: "a" });
    const b = store.addLog({ message: "b" });
    const c = store.addLog({ message: "c" });
    expect(a._id).toBe(0);
    expect(b._id).toBe(1);
    expect(c._id).toBe(2);
  });

  it("drops the oldest entry once maxLogs is exceeded (ring buffer)", () => {
    const store = new ProjectLogsStore(3);
    store.addLog({ message: "a" });
    store.addLog({ message: "b" });
    store.addLog({ message: "c" });
    store.addLog({ message: "d" });
    const messages = store.getAll().map((e) => e.message);
    expect(messages).toEqual(["b", "c", "d"]);
  });

  it("continues to assign fresh _ids even after buffer wraps", () => {
    const store = new ProjectLogsStore(2);
    store.addLog({ message: "a" });
    store.addLog({ message: "b" });
    const third = store.addLog({ message: "c" });
    expect(third._id).toBe(2);
    expect(store.getAll().map((e) => e._id)).toEqual([1, 2]);
  });

  it("filters by level", () => {
    const store = new ProjectLogsStore(100);
    store.addLog({ message: "a", level: V1LogLevel.LOG_LEVEL_INFO });
    store.addLog({ message: "b", level: V1LogLevel.LOG_LEVEL_ERROR });
    store.addLog({ message: "c", level: V1LogLevel.LOG_LEVEL_WARN });

    const errors = store.getFiltered({
      levels: [V1LogLevel.LOG_LEVEL_ERROR],
    });
    expect(errors.map((e) => e.message)).toEqual(["b"]);
  });

  it("search matches message and jsonPayload case-insensitively", () => {
    const store = new ProjectLogsStore(100);
    store.addLog({ message: "Fooooo", jsonPayload: "{}" });
    store.addLog({ message: "bar", jsonPayload: '{"reason":"foobar"}' });
    store.addLog({ message: "baz", jsonPayload: "{}" });

    const hits = store.getFiltered({ search: "FOO" });
    expect(hits.map((e) => e.message)).toEqual(["Fooooo", "bar"]);
  });

  it("combines level and search as an intersection", () => {
    const store = new ProjectLogsStore(100);
    store.addLog({
      message: "boot",
      level: V1LogLevel.LOG_LEVEL_ERROR,
    });
    store.addLog({
      message: "boot",
      level: V1LogLevel.LOG_LEVEL_INFO,
    });
    store.addLog({
      message: "crash",
      level: V1LogLevel.LOG_LEVEL_ERROR,
    });

    const hits = store.getFiltered({
      levels: [V1LogLevel.LOG_LEVEL_ERROR],
      search: "boot",
    });
    expect(hits.map((e) => e.message)).toEqual(["boot"]);
  });

  it("returns everything when filters are empty", () => {
    const store = new ProjectLogsStore(100);
    store.addLog({ message: "a" });
    store.addLog({ message: "b" });
    expect(store.getFiltered({}).map((e) => e.message)).toEqual(["a", "b"]);
    expect(store.getFiltered().map((e) => e.message)).toEqual(["a", "b"]);
  });
});
