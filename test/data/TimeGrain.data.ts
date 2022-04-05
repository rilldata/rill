import type { DataProviderData } from "@adityahegde/typescript-test-utils";
import { TimeGrain } from "$common/database-service/DatabaseColumnActions";

export interface GeneratedTimeseriesTestCase {
    start: string,
    end: string,
    interval: string,
    table: string,
    expectedTimeGrain: string
}

export const timeGrainSeriesData:DataProviderData<[GeneratedTimeseriesTestCase]> = { subData: [
    {
        title: 'different ms, same day',
        subData: [{ args: [{
                table:'ts_ms_01',
                start: '2021-01-01 01:30:00',
                end: '2021-01-01 02:00:00',
                interval: '4 millisecond',
                expectedTimeGrain: TimeGrain.ms
            }]}
        ]
    },
    {
        title: 'ms but over the date border',
        subData: [{ args: [{
                table:'ts_ms_03',
                start: '2021-01-01 23:30:00',
                end: '2021-01-02 23:30:00',
                interval: '4 millisecond',
                expectedTimeGrain: TimeGrain.ms
            }]}]
    },
    {
        title: 'ts_seconds_01',
        subData: [{ args: [{
            table:'ts_seconds_01',
            start: '2021-01-01 12:30:00',
            end: '2021-01-01 13:30:00',
            interval: '2 seconds',
            expectedTimeGrain: TimeGrain.second
        }]}]
    },
    {
        title: "ts_minutes_01",
        subData: [{ args: [{
            table:'ts_minutes_01',
            start: '2021-01-01 01:04:04',
            end: '2021-01-01 09:04:04',
            interval: '47 minutes',
            expectedTimeGrain: TimeGrain.minute
        }]}]
    },
    {
        title: 'ts_hours_01',
        subData: [{ args: [{
            table:'ts_hours_01',
            start: '2021-01-01',
            end: '2022-01-01',
            interval: '2 hours',
            expectedTimeGrain: TimeGrain.hour
        }]}]
    },
    {
        title: 'ts_days_01',
        subData: [{ args: [{
            table:'ts_days_01',
            start: '2021-01-01',
            end: '2025-01-01',
            interval: '3 days',
            expectedTimeGrain: TimeGrain.day
        }]}]
    },
    {
        title: 'ts_weeks_01',
        subData: [{ args: [{
            table:'ts_weeks_01',
            start: '1900-01-01',
            end: '2000-01-01',
            interval: '7 day',
            expectedTimeGrain: TimeGrain.week
        }]}]
    },
    {
        title: 'ts_weeks_02',
        subData: [{ args: [{
            table:'ts_weeks_02',
            start: '1900-01-01',
            end: '1900-03-01',
            interval: '7 day',
            expectedTimeGrain: TimeGrain.week
        }]}]
    },
    {
        title: "weekly, 100 years",
        subData: [{ args: [{
            table:'ts_weeks_03',
            start: '1900-01-01',
            end: '2000-01-01',
            interval: '7 day',
            expectedTimeGrain: TimeGrain.week
        }]}]
    },
    {
        title: 'ts_years_01',
        subData: [{ args: [{
            table:'ts_years_01',
            start: '1900-01-01',
            end: '2000-01-01',
            interval: '1 year',
            expectedTimeGrain: TimeGrain.year
        }]}]
    },
]}