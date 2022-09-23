import type { EntityRecord } from "../../../common/data-modeler-state-service/entity-state-service/EntityStateService";

export interface ValidationConfig<Entity extends EntityRecord> {
  field: keyof Entity;
  // takes the existing entity and current changes
  // returns changes from a validation endpoint
  validate: (
    entity: Entity,
    changes: Partial<Entity>
  ) => Promise<Partial<Entity>>;
  validationPassed?: (changes: Partial<Entity>) => boolean;
}

export async function validateEntity<Entity extends EntityRecord>(
  entity: Entity,
  changes: Partial<Entity>,
  validations: Array<ValidationConfig<Entity>>
): Promise<Partial<Entity>> {
  let validationChanges: Partial<Entity> = {};
  await Promise.all(
    validations
      .filter((validation) => validation.field in changes)
      .map(async (validation) => {
        validationChanges = {
          ...(await validation.validate(entity, changes)),
          ...validationChanges,
        };
      })
  );
  return validationChanges;
}

export function validationSucceeded<Entity extends EntityRecord>(
  changes: Partial<Entity>,
  validations: Array<ValidationConfig<Entity>>
) {
  return validations.every(
    (validation) =>
      !validation.validationPassed || validation.validationPassed(changes)
  );
}
