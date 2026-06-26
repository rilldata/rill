import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { EnvStore } from "@rilldata/web-common/features/env-management/env-store.ts";

// fileArtifacts.getFileArtifact("/.env") drives the getter. Each test sets
// `mockEnvContent`, then triggers `envStore.pull()` to exercise the parser.
let mockEnvContent: string | undefined = undefined;
const fetchContent = vi.fn(async () => mockEnvContent);

vi.mock(
  "@rilldata/web-common/features/entity-management/file-artifacts",
  () => ({
    fileArtifacts: {
      getFileArtifact: vi.fn(() => ({ fetchContent })),
    },
  }),
);

vi.mock(
  "@rilldata/web-common/features/entity-management/edit-environment.ts",
  () => ({
    isCloudRuntimeEditEnvironment: vi.fn(() => false),
  }),
);

vi.mock("@rilldata/web-common/runtime-client", () => ({
  runtimeServicePutFile: vi.fn(async () => ({})),
  runtimeServicePushEnv: vi.fn(async () => ({})),
}));

// createEnvFileStore calls setContext, which only works inside a Svelte
// component lifecycle. Replace it with a simple in-memory slot so we can
// retrieve the store via getEnvFileStore the same way callers do.
let capturedStore: EnvStore | null = null;
vi.mock("svelte", () => ({
  setContext: vi.fn((_key: string, value: EnvStore) => {
    capturedStore = value;
  }),
  getContext: vi.fn(() => capturedStore),
}));

import {
  runtimeServicePushEnv,
  runtimeServicePutFile,
} from "@rilldata/web-common/runtime-client";
import { isCloudRuntimeEditEnvironment } from "@rilldata/web-common/features/entity-management/edit-environment.ts";
import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
import { EnvEditSession } from "@rilldata/web-common/features/env-management/env-edit-session.ts";
import { createEnvFileStore, getEnvFileStore } from "./env-file-store";

const runtimeClient = { instanceId: "inst-1" } as never;

async function setupStore(envContent: string | undefined) {
  mockEnvContent = envContent;
  const unsubscribe = createEnvFileStore(runtimeClient);
  const store = getEnvFileStore();
  // createEnvFileStore fires `void envStore.pull()`; await an explicit pull
  // so assertions don't race with that floating promise.
  await store.pull();
  return { store, unsubscribe };
}

describe("env-file-store", () => {
  beforeEach(() => {
    capturedStore = null;
    mockEnvContent = undefined;
    vi.mocked(isCloudRuntimeEditEnvironment).mockReturnValue(false);
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe("pull / getter", () => {
    it("parses basic KEY=VALUE pairs", async () => {
      const { store, unsubscribe } = await setupStore("FOO=bar\nBAZ=qux");
      expect(store.store.get("FOO")?.value).toBe("bar");
      expect(store.store.get("BAZ")?.value).toBe("qux");
      unsubscribe();
    });

    it("strips empty and whitespace-only lines", async () => {
      const { store, unsubscribe } = await setupStore(
        "\nFOO=bar\n   \n\nBAZ=qux\n",
      );
      expect([...store.store.keys()]).toEqual(["FOO", "BAZ"]);
      unsubscribe();
    });

    it("strips comment-only lines", async () => {
      const { store, unsubscribe } = await setupStore(
        "# leading comment\nFOO=bar\n  # indented comment\nBAZ=qux",
      );
      expect([...store.store.keys()]).toEqual(["FOO", "BAZ"]);
      unsubscribe();
    });

    it("retains '#' inside quoted values", async () => {
      const { store, unsubscribe } = await setupStore(
        `URL="https://example.com#section"\nPASSWORD='foo#bar'`,
      );
      expect(store.store.get("URL")?.value).toBe("https://example.com#section");
      expect(store.store.get("PASSWORD")?.value).toBe("foo#bar");
      unsubscribe();
    });

    it("preserves '=' inside values", async () => {
      const { store, unsubscribe } = await setupStore(
        "DSN=postgres://u:p@h/db?opt=1&other=2",
      );
      expect(store.store.get("DSN")?.value).toBe(
        "postgres://u:p@h/db?opt=1&other=2",
      );
      unsubscribe();
    });

    it("returns an empty store when the .env file is missing", async () => {
      const { store, unsubscribe } = await setupStore(undefined);
      expect(store.store.size).toBe(0);
      unsubscribe();
    });

    it("returns an empty store when the .env file is empty", async () => {
      const { store, unsubscribe } = await setupStore("");
      expect(store.store.size).toBe(0);
      unsubscribe();
    });
  });

  describe("flush / setter", () => {
    it("writes key=value lines via runtimeServicePutFile", async () => {
      const { store, unsubscribe } = await setupStore("");
      const session = new EnvEditSession(store, "clickhouse");
      session.acquire("password", "secret", "CLICKHOUSE_PASSWORD");
      await session.commit();

      expect(runtimeServicePutFile).toHaveBeenCalledWith(runtimeClient, {
        path: "/.env",
        blob: "CLICKHOUSE_PASSWORD=secret",
        create: true,
        createOnly: false,
      });
      unsubscribe();
    });

    it("calls runtimeServicePushEnv on cloud", async () => {
      vi.mocked(isCloudRuntimeEditEnvironment).mockReturnValue(true);
      const { store, unsubscribe } = await setupStore("");
      const session = new EnvEditSession(store, "clickhouse");
      session.acquire("password", "secret", "CLICKHOUSE_PASSWORD");
      await session.commit();

      expect(runtimeServicePushEnv).toHaveBeenCalledWith(runtimeClient, {});
      unsubscribe();
    });

    it("skips runtimeServicePushEnv when not on cloud", async () => {
      const { store, unsubscribe } = await setupStore("");
      const session = new EnvEditSession(store, "clickhouse");
      session.acquire("password", "secret", "CLICKHOUSE_PASSWORD");
      await session.commit();

      expect(runtimeServicePushEnv).not.toHaveBeenCalled();
      unsubscribe();
    });
  });

  describe("event-bus subscription", () => {
    it("re-pulls when an env-file-updated event fires", async () => {
      mockEnvContent = "FOO=initial";
      const unsubscribe = createEnvFileStore(runtimeClient);
      const store = getEnvFileStore();
      await store.pull();
      expect(store.store.get("FOO")?.value).toBe("initial");

      mockEnvContent = "FOO=updated";
      eventBus.emit("env-file-updated", "/.env");
      // The handler calls `void envStore.pull()`; wait for the floating
      // promise to settle before asserting.
      await new Promise((r) => setTimeout(r, 0));

      expect(store.store.get("FOO")?.value).toBe("updated");
      unsubscribe();
    });

    it("stops re-pulling after the returned unsubscribe is called", async () => {
      mockEnvContent = "FOO=initial";
      const unsubscribe = createEnvFileStore(runtimeClient);
      const store = getEnvFileStore();
      await store.pull();
      unsubscribe();

      mockEnvContent = "FOO=updated";
      eventBus.emit("env-file-updated", "/.env");
      await new Promise((r) => setTimeout(r, 0));

      expect(store.store.get("FOO")?.value).toBe("initial");
    });
  });
});
