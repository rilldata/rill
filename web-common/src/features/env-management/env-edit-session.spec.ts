import { describe, expect, it } from "vitest";
import { EnvEditSession } from "@rilldata/web-common/features/env-management/env-edit-session.ts";
import { makeTestEnvStore } from "@rilldata/web-common/features/env-management/test/test-env-store.ts";

describe("EnvEditSession", () => {
  it("reuses the same mapped name for a key across preview edits", async () => {
    const { envStore } = await makeTestEnvStore();
    const session = new EnvEditSession(envStore, "ns");

    session.startEdit();
    const first = session.acquire("alpha", "v1");
    expect(first.mappedEnvVarName).toBe("NS_ALPHA");

    // A subsequent preview (no commit in between) must reuse the allocation.
    session.startEdit();
    const second = session.acquire("alpha", "v2");
    expect(second.mappedEnvVarName).toBe("NS_ALPHA");
  });

  it("suffixes distinct keys that resolve to the same generic name", async () => {
    const { envStore } = await makeTestEnvStore();
    const session = new EnvEditSession(envStore, "ns");

    session.startEdit();
    // Two different keys whose generic names collide ("a-b" and "a.b" both
    // normalise to NS_A_B).
    const first = session.acquire("a-b", "v1");
    const second = session.acquire("a.b", "v2");
    expect(first.mappedEnvVarName).toBe("NS_A_B");
    expect(second.mappedEnvVarName).toBe("NS_A_B_1");
  });

  // A name held by an in-flight (uncommitted preview) entry must not be reassigned
  // to a new key acquired in the same edit. The explicit-envVarName path is the one that
  // bypasses key-to-name normalisation, so it can drive two distinct keys onto
  // the same requested name.
  it("does not reassign an in-flight name to a new key", async () => {
    const { testEnvs, envStore } = await makeTestEnvStore();
    const session = new EnvEditSession(envStore, "ns");

    // Preview 1: "alpha" claims SHARED. No commit, so it stays in-flight.
    session.startEdit();
    const alpha1 = session.acquire("alpha", "v1", "SHARED");
    expect(alpha1.mappedEnvVarName).toBe("SHARED");

    // Preview 2: a new key "beta" also requests SHARED, then "alpha" reclaims
    // its allocation. The two must not collapse onto the same mapped name.
    session.startEdit();
    const beta = session.acquire("beta", "v2", "SHARED");
    const alpha2 = session.acquire("alpha", "v1b", "SHARED");

    expect(beta.mappedEnvVarName).toBe("SHARED_1");
    expect(alpha2.mappedEnvVarName).toBe("SHARED");
    expect(beta.mappedEnvVarName).not.toBe(alpha2.mappedEnvVarName);

    // And both values must survive the flush — neither overwrites the other.
    await session.commit();
    expect(testEnvs).toEqual({
      SHARED: "v1b",
      SHARED_1: "v2",
    });
  });

  it("avoids names already present in the parent store", async () => {
    const { envStore } = await makeTestEnvStore({ NS_ALPHA: "external" });
    const session = new EnvEditSession(envStore, "ns");

    session.startEdit();
    const entry = session.acquire("alpha", "v1");
    expect(entry.mappedEnvVarName).toBe("NS_ALPHA_1");
  });
});
