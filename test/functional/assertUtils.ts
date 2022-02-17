import type { ProfileColumn } from "$lib/types";
import type { TestDataColumns } from "../data/DataLoader.data";

export function assertColumns(profileColumns: ProfileColumn[], columns: TestDataColumns): void {
    profileColumns.forEach((profileColumn, idx) => {
        expect(profileColumn.name).toBe(columns[idx].name);
        expect(profileColumn.type).toBe(columns[idx].type);
        expect(profileColumn.nullCount > 0).toBe(columns[idx].isNull);
        // TODO: assert summary
        // console.log(profileColumn.name, profileColumn.summary);
    });
}
