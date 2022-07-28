import type { EntityState, PayloadAction } from "@reduxjs/toolkit";

type IdRecordWithField<Key extends string | number | symbol, Value> = {
  id: string;
} & Record<Key, Value>;

export const setFieldPrepare =
  <Entity, Key extends keyof Entity>(key: Key) =>
  (id: string, value: Entity[Key]) =>
    ({
      payload: { id, [key]: value },
    } as PayloadAction<IdRecordWithField<Key, Entity[Key]>>);
export const setFieldReducer =
  <Entity, Key extends keyof Entity>(key: Key) =>
  (
    state: EntityState<Entity>,
    { payload }: PayloadAction<IdRecordWithField<Key, Entity[Key]>>
  ) => {
    if (!state.entities[payload.id]) return;
    state.entities[payload.id][key] = payload[key] as unknown as Entity[Key];
  };
