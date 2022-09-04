import { EntityRepository } from "$common/data-modeler-state-service/sync-service/EntityRepository";
import type { PersistentModelEntity } from "$common/data-modeler-state-service/entity-state-service/PersistentModelEntityService";
import type { StateConfig } from "$common/config/StateConfig";
import type {
  EntityType,
  StateType,
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import {
  existsSync,
  mkdirSync,
  readFileSync,
  writeFileSync,
  statSync,
  readdirSync,
  unlinkSync,
} from "fs";
import type { EntityState } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type { DataModelerService } from "$common/data-modeler-service/DataModelerService";
import { Throttler } from "$common/utils/Throttler";
import path from "path";

export class PersistentModelRepository extends EntityRepository<PersistentModelEntity> {
  private readonly saveDirectory: string;
  private filesForEntities = new Map<string, string>();
  private previousFilesForEntities: Map<string, string>;
  private throttler = new Throttler();

  constructor(
    stateConfig: StateConfig,
    dataModelerService: DataModelerService,
    entityType: EntityType,
    stateType: StateType
  ) {
    super(stateConfig, dataModelerService, entityType, stateType);
    this.saveDirectory = stateConfig.modelFolder;
    if (!existsSync(this.saveDirectory)) {
      mkdirSync(this.saveDirectory, { recursive: true });
    }
  }

  /**
   * Persist the entity query to file.
   */
  public async save(entity: PersistentModelEntity): Promise<void> {
    writeFileSync(
      `${this.saveDirectory}/${this.getFileName(entity)}`,
      this.getFileContent(entity)
    );
  }

  public async getAll(): Promise<EntityState<PersistentModelEntity>> {
    const currentFiles = new Map<string, string>();
    // store ids of previous entity for the file
    // if it is a new file then store an empty string
    readdirSync(this.saveDirectory).forEach((file) =>
      currentFiles.set(file, this.filesForEntities?.get(file) ?? "")
    );
    this.previousFilesForEntities = this.filesForEntities;
    this.filesForEntities = new Map<string, string>();

    const entityState = await super.getAll();

    // build the new map of file name to entity id
    entityState.entities.forEach((entity) =>
      this.filesForEntities.set(this.getFileName(entity), entity.id)
    );

    const filesToBeRemoved = new Map<string, string>();

    // go through all current files.
    currentFiles.forEach((id, currentFile) => {
      // if a file has an entry in new map ignore it.
      if (
        this.filesForEntities.has(currentFile) ||
        !this.isValidFile(currentFile)
      )
        return;
      const currentFilePath = `${this.saveDirectory}/${currentFile}`;

      if (id) {
        // if the file has no entry in new map but has an id in currentFiles then remove the file.
        // this was a possible rename
        filesToBeRemoved.set(
          path.parse(currentFile).name.toLowerCase(),
          currentFilePath
        );
      } else {
        // else this a file added from outside.
        // create a new model
        setTimeout(() => {
          // throttle the call
          this.throttler.throttle(
            currentFile,
            () => {
              this.createEntity(
                currentFile,
                readFileSync(currentFilePath).toString()
              );
            },
            this.stateConfig.syncInterval * 2
          );
          // add a small timeout to make sure it runs after the sync ends
        }, 5);
      }
    });
    this.filesForEntities.forEach((id, fileName) => {
      if (currentFiles.has(fileName)) return;
      const normalisedEntityName = path.parse(fileName).name.toLowerCase();
      if (this.previousFilesForEntities.get(fileName) === id) {
        // if current files is missing one of the entities then it is a possible delete.
        setTimeout(() => {
          // throttle the call
          this.throttler.throttle(
            id,
            () => {
              this.deleteEntity(id);
            },
            this.stateConfig.syncInterval * 2
          );
          // add a small timeout to make sure it runs after the sync ends
        }, 5);
      } else if (filesToBeRemoved.has(normalisedEntityName)) {
        this.renameFileToDifferentCase(
          filesToBeRemoved.get(normalisedEntityName),
          `${this.saveDirectory}/${fileName}`
        );
        filesToBeRemoved.delete(normalisedEntityName);
      }
    });
    for (const fileToBeRemoved of filesToBeRemoved.values()) {
      unlinkSync(fileToBeRemoved);
    }

    return entityState;
  }

  /**
   * Update specific fields in entity based on id or any other field
   */
  public async update(entity: PersistentModelEntity): Promise<boolean> {
    const modelFileName = this.getFileName(entity);
    const modelFilePath = `${this.saveDirectory}/${modelFileName}`;
    if (!existsSync(modelFilePath)) {
      // call save for fresh entities
      if (
        !this.previousFilesForEntities.has(modelFileName) ||
        this.previousFilesForEntities.get(modelFileName) !== entity.id
      ) {
        await this.save(entity);
      }
      return false;
    }

    const newQuery = readFileSync(modelFilePath).toString();
    const fileUpdated = statSync(modelFilePath).mtimeMs;
    if (
      this.contentHasChanged(entity, newQuery) &&
      fileUpdated > entity.lastUpdated
    ) {
      this.updateEntity(entity, newQuery);
      entity.lastUpdated = Date.now();
      return true;
    }
    return false;
  }

  // adding some abstraction in anticipation of other entities persisting to files
  protected getFileName(entity: PersistentModelEntity): string {
    return entity.name;
  }
  protected getFileContent(entity: PersistentModelEntity): string {
    return entity.query;
  }
  protected updateEntity(
    entity: PersistentModelEntity,
    newContent: string
  ): void {
    entity.query = newContent;
  }
  protected contentHasChanged(
    entity: PersistentModelEntity,
    newContent: string
  ): boolean {
    return newContent !== entity.query;
  }
  protected isValidFile(fileName: string): boolean {
    return fileName.endsWith(".sql");
  }

  protected async createEntity(fileName: string, fileContent: string) {
    return this.dataModelerService.dispatch("addModel", [
      { name: fileName, query: fileContent },
    ]);
  }

  protected async deleteEntity(id: string) {
    return this.dataModelerService.dispatch("deleteModel", [id]);
  }

  /**
   * Treats file systems as case-sensitive to match model name.
   * Removes old file and re adds new one to get around this.
   */
  private renameFileToDifferentCase(fromFileName: string, toFileName: string) {
    const content = readFileSync(fromFileName).toString();
    unlinkSync(fromFileName);
    writeFileSync(toFileName, content);
  }
}
