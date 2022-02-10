import {DataGeneratorType, ParquetDataType} from "./DataGeneratorType";
import {CITY_NULL_CHANCE, LOCATIONS, MAX_USERS} from "./data-constants";

export interface User {
    id: number;
    name: string;
    city?: string;
    country: string;
}

export class UsersGeneratorType extends DataGeneratorType {
    public generateRow(id: number): User {
        if (id >= MAX_USERS) return null;
        const location = this.selectRandomEntry(LOCATIONS);
        const user: User = {
            id,
            name: `User${id}`,
            country: location[1],
        };
        if (Math.random() > CITY_NULL_CHANCE) {
            user.city = location[0];
        }
        return user;
    }

    public getParquetSchema(): Record<keyof User, ParquetDataType> {
        return {
            id: { type: "INT64" },
            name: { type: "UTF8" },
            city: { type: "UTF8", optional: true },
            country: { type: "UTF8" },
        };
    }
}
