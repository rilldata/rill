import {DataGeneratorType, ParquetDataType} from "./DataGeneratorType";

const PUBLISHER_DOMAINS = [
    ["Yahoo", "sports.yahoo.com"],
    ["Yahoo", "news.yahoo.com"],
    ["Microsoft", "msn.com"],
    ["Google", "news.google.com"],
    ["Google", "google.com"],
    ["Facebook", "facebook.com"],
    ["Facebook", "instagram.com"],
];
const START_DATE = "2022-01-01";
const END_DATE = "2022-01-31";
const BID_START = 1;
const BID_END = 2;

export interface AdBid {
    id: number;
    timestamp: number;
    publisher: string;
    domain: string;
    bid_price: number;
}

export class AdBidsGeneratorType extends DataGeneratorType {
    public generateRow(id: number): AdBid {
        const publisherDomain = this.selectRandomEntry(PUBLISHER_DOMAINS);
        return {
            id,
            timestamp: this.generateRandomTimestamp(START_DATE, END_DATE),
            publisher: publisherDomain[0],
            domain: publisherDomain[1],
            bid_price: this.generateRandomFloat(BID_START, BID_END),
        };
    }

    public getParquetSchema(): Record<keyof AdBid, ParquetDataType> {
        return {
            id: { type: "INT64" },
            timestamp: { type: "TIMESTAMP_MILLIS" },
            publisher: { type: "UTF8" },
            domain: { type: "UTF8" },
            bid_price: { type: "DOUBLE" },
        };
    }
}
