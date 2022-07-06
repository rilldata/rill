import { DataGeneratorType, ParquetDataType } from "./DataGeneratorType";
import {
  BID_END,
  BID_START,
  END_DATE,
  PUBLISHER_DOMAINS,
  PUBLISHER_DOMAINS_BY_MONTH,
  PUBLISHER_NULL_CHANCE,
  START_DATE,
} from "../data-constants";

export interface AdBid {
  id: number;
  timestamp: string;
  publisher?: string;
  domain: string;
  bid_price: number;
}

export class AdBidsGeneratorType extends DataGeneratorType {
  public columnsOrder = ["id", "timestamp", "publisher", "domain", "bid_price"];

  public generateRow(id: number): AdBid {
    const pubFactor = Math.random() > 0.5;

    const timestamp = this.generateRandomTimestamp(START_DATE, END_DATE);
    const month = new Date(timestamp).getMonth();
    const publisherDomain = this.selectRandomEntry(
      pubFactor ? PUBLISHER_DOMAINS_BY_MONTH[month] : PUBLISHER_DOMAINS
    );

    const adBid: AdBid = {
      id,
      timestamp,
      domain: publisherDomain[1],
      bid_price: this.generateRandomFloat(
        BID_START + (pubFactor ? 2 : 0),
        BID_END + (pubFactor ? 4 : 0)
      ),
    };

    if (Math.random() > PUBLISHER_NULL_CHANCE) {
      adBid.publisher = publisherDomain[0];
    }
    return adBid;
  }

  public getParquetSchema(): Record<keyof AdBid, ParquetDataType> {
    return {
      id: { type: "INT32" },
      timestamp: { type: "UTF8" },
      publisher: { type: "UTF8", optional: true },
      domain: { type: "UTF8" },
      bid_price: { type: "DOUBLE" },
    };
  }
}
