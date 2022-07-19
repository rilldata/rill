import { DataGeneratorType, ParquetDataType } from "./DataGeneratorType";
import {
  CITY_NULL_CHANCE,
  LOCATIONS,
  MAX_USERS,
  USER_NULL_CHANCE,
} from "../data-constants";

export interface AdImpression {
  id: number;
  city?: string;
  country: string;
  user_id?: number;
}

export class AdImpressionsGeneratorType extends DataGeneratorType {
  public csvExtension = "tsv";
  public csvDelimiter = "\t";
  public columnsOrder = ["id", "city", "country", "user_id"];

  public generateRow(id: number): AdImpression {
    const location = this.selectRandomEntry(LOCATIONS);
    const adImpression: AdImpression = {
      id: id * 2,
      country: location[1],
    };
    if (Math.random() > CITY_NULL_CHANCE) {
      adImpression.city = location[0];
    }
    if (Math.random() > USER_NULL_CHANCE) {
      adImpression.user_id = this.generateRandomInt(0, MAX_USERS - 1);
    }
    return adImpression;
  }

  public getParquetSchema(): Record<keyof AdImpression, ParquetDataType> {
    return {
      id: { type: "INT32" },
      city: { type: "UTF8", optional: true },
      country: { type: "UTF8" },
      user_id: { type: "INT32", optional: true },
    };
  }
}
