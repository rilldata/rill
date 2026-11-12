import type { Query } from "@tanstack/svelte-query";
import { afterEach, describe, expect, it, vi } from "vitest";
import { createQueryClient, onQueryError } from "./globalQueryClient";

describe("onQueryError", () => {
  const removers: (() => void)[] = [];
  afterEach(() => removers.splice(0).forEach((remove) => remove()));

  async function failQuery(key: string) {
    const error = new Error("boom");
    await expect(
      createQueryClient().fetchQuery({
        queryKey: [key],
        queryFn: () => Promise.reject(error),
      }),
    ).rejects.toBe(error);
    return error;
  }

  it("calls every listener with the error and query of a failed query", async () => {
    // Mirrors the root layouts' telemetry handler, which needs the query.
    const appHandler = vi.fn((_error: unknown, query: Query) => query.queryKey);
    const signInListener = vi.fn();
    removers.push(onQueryError(appHandler), onQueryError(signInListener));

    const error = await failQuery("failing");

    expect(appHandler).toHaveBeenCalledOnce();
    expect(appHandler.mock.calls[0][0]).toBe(error);
    expect(appHandler.mock.calls[0][1].queryKey).toEqual(["failing"]);
    expect(signInListener).toHaveBeenCalledOnce();
    expect(signInListener.mock.calls[0][0]).toBe(error);
  });

  it("stops calling a removed listener", async () => {
    const listener = vi.fn();
    onQueryError(listener)();

    await failQuery("removed");

    expect(listener).not.toHaveBeenCalled();
  });
});
