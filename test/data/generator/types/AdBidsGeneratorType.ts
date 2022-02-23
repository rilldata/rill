import {DataGeneratorType, ParquetDataType} from "./DataGeneratorType";
import {BID_END, BID_START, END_DATE, PUBLISHER_DOMAINS, PUBLISHER_NULL_CHANCE, START_DATE} from "../data-constants";

export interface AdBid {
    id: number;
    timestamp: number;
    publisher?: string;
    domain: string;
    bid_price: number;
}

export class AdBidsGeneratorType extends DataGeneratorType {
    public columnsOrder = ["id", "timestamp", "publisher", "domain", "bid_price"];

    public generateRow(id: number): AdBid {
        const publisherDomain = this.selectRandomEntry(PUBLISHER_DOMAINS);
        const adBid: AdBid = {
            id,
            timestamp: this.generateRandomTimestamp(START_DATE, END_DATE),
            domain: publisherDomain[1],
            bid_price: this.generateRandomFloat(BID_START, BID_END),
        };
        if (Math.random() > PUBLISHER_NULL_CHANCE) {
            adBid.publisher = publisherDomain[0];
        }
        return adBid;
    }

    public getParquetSchema(): Record<keyof AdBid, ParquetDataType> {
        return {
            id: { type: "INT32" },
            timestamp: { type: "TIMESTAMP_MILLIS" },
            publisher: { type: "UTF8", optional: true },
            domain: { type: "UTF8" },
            bid_price: { type: "DOUBLE" },
        };
    }
}
