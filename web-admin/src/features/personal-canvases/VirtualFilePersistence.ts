import {
  adminServiceEditPersonalVirtualFile,
  adminServiceGetPersonalVirtualFile,
} from "@rilldata/web-admin/client";
import { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";

export type PersonalVirtualFileType = "PERSONAL_VIRTUAL_FILE_TYPE_CANVAS";

export interface VirtualFilePersistenceOptions {
  org: string;
  project: string;
  type: PersonalVirtualFileType;
  /**
   * Personal virtual file resource name (the URL-safe slug used by the admin API).
   */
  name: string;
  /**
   * Display name. Used by the editor host to decorate the UI; not required for save.
   */
  displayName?: string;
}

/**
 * VirtualFilePersistence reuses FileArtifact for store + lifecycle plumbing, but redirects
 * the read and write transports to the admin server's personal virtual file RPCs.
 *
 * The "path" parameter exists only as a cache key inside FileArtifact. We synthesize a
 * stable virtual key from (org, project, type, name) so multiple personal canvases coexist
 * without colliding in the query cache.
 */
export class VirtualFilePersistence extends FileArtifact {
  private readonly org: string;
  private readonly project: string;
  private readonly type: PersonalVirtualFileType;
  private readonly canvasName: string;
  readonly displayName: string | undefined;

  constructor(client: RuntimeClient, opts: VirtualFilePersistenceOptions) {
    // Use a virtual path that won't collide with any real project file path.
    const virtualPath = `__personal__/${opts.org}/${opts.project}/${opts.type}/${opts.name}.yaml`;
    super(client, virtualPath);
    this.org = opts.org;
    this.project = opts.project;
    this.type = opts.type;
    this.canvasName = opts.name;
    this.displayName = opts.displayName;

    // Bind the runtime resource upfront so getResource()/getParseError() can subscribe to
    // the live canvas resource without waiting for inference from the YAML body.
    const kind = personalVirtualFileTypeToResourceKind(opts.type);
    if (kind) {
      this.resourceName.set({ kind, name: opts.name });
      this.inferredResourceKind.set(kind);
    }
  }

  // Read canonical YAML from the admin server instead of the runtime file repo.
  protected async fetchBlob(_invalidate: boolean): Promise<string | undefined> {
    try {
      const response = await adminServiceGetPersonalVirtualFile(
        this.org,
        this.project,
        this.type,
        this.canvasName,
      );
      return response.yaml ?? "";
    } catch (e) {
      console.error("VirtualFilePersistence.fetchBlob failed", e);
      return undefined;
    }
  }

  // Persist edits via the admin EditPersonalVirtualFile RPC. The runtime continues to
  // surface the resulting catalog resource for rendering once the admin server triggers
  // reconcile.
  protected async putBlob(blob: string): Promise<void> {
    await adminServiceEditPersonalVirtualFile(
      this.org,
      this.project,
      this.type,
      this.canvasName,
      { yaml: blob },
    );
  }
}

function personalVirtualFileTypeToResourceKind(
  type: PersonalVirtualFileType,
): ResourceKind | null {
  switch (type) {
    case "PERSONAL_VIRTUAL_FILE_TYPE_CANVAS":
      return ResourceKind.Canvas;
    default:
      return null;
  }
}
