import type {
  CanvasComponent,
  ComponentSize,
} from "@rilldata/web-common/features/canvas/components/types";
import { getParsedDocument } from "@rilldata/web-common/features/canvas/inspector/selectors";
import type { ComponentInputParam } from "@rilldata/web-common/features/canvas/inspector/types";
import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
import { get, writable, type Writable } from "svelte/store";

// A base class that implements all the store logic
export abstract class BaseCanvasComponent<T> implements CanvasComponent<T> {
  specStore: Writable<T>;
  pathInYAML: (string | number)[] = [];
  fileArtifact: FileArtifact;

  // Let child classes define these
  abstract minSize: ComponentSize;
  abstract defaultSize: ComponentSize;
  abstract isValid(spec: T): boolean;
  abstract inputParams(): Record<keyof T, ComponentInputParam>;

  constructor(
    fileArtifact: FileArtifact,
    path: (string | number)[],
    defaultSpec: T,
    initialSpec: Partial<T> = {},
  ) {
    // Initialize the store with merged spec
    const mergedSpec = { ...defaultSpec, ...initialSpec };
    this.specStore = writable(mergedSpec);
    this.pathInYAML = path;
    this.fileArtifact = fileArtifact;
  }

  private async updateYAML(newSpec: T): Promise<void> {
    const parseDocumentStore = getParsedDocument(this.fileArtifact);
    const parsedDocument = get(parseDocumentStore);

    const { saveContent } = this.fileArtifact;

    // Update the Item
    parsedDocument.setIn(this.pathInYAML, newSpec);

    // Save the updated document
    await saveContent(parsedDocument.toString());
  }

  async setSpec(newSpec: T): Promise<void> {
    if (this.isValid(newSpec)) {
      await this.updateYAML(newSpec);
    }
    this.specStore.set(newSpec);
  }

  // TODO: Add stricter type definition for keys and value deriving from spec
  async updateProperty(key: string, value: unknown): Promise<void> {
    const currentSpec = get(this.specStore);
    const newSpec = { ...currentSpec, [key]: value };
    if (this.isValid(newSpec)) {
      await this.updateYAML(newSpec);
    }
    this.specStore.set(newSpec);
  }
}
