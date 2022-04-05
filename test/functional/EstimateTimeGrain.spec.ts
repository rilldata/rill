import { DataProviderData, TestBase } from "@adityahegde/typescript-test-utils";
import { JestTestLibrary } from "@adityahegde/typescript-test-utils/dist/jest/JestTestLibrary";
import { FunctionalTestBase } from "./FunctionalTestBase";
import type { DatabaseActionsDefinition, DatabaseService } from "$common/database-service/DatabaseService";
import { RootConfig } from "$common/config/RootConfig";
import { DatabaseConfig } from "$common/config/DatabaseConfig";
import { StateConfig } from "$common/config/StateConfig";
import { dataModelerServiceFactory } from "$common/serverFactory";

interface GeneratedTimeseriesTestCase {
    start: string,
    end: string,
    interval: string,
    table: string,
    expectedTimeGrain: string
}

const SYNC_TEST_FOLDER = "temp/sync-test";

function ctas(table, select_statement, temp = true) {
    return `CREATE ${temp ? 'TEMPORARY' : ''} VIEW ${table} AS (${select_statement})`
}

function generateSeries(table:string, start:string, end:string, interval:string) {
    return ctas(table, `SELECT generate_series as ts from generate_series(TIMESTAMP '${start}', TIMESTAMP '${end}', interval ${interval})`)
}

const subData = { subData: [
    {
        title: 'different ms, same day',
        subData: [{ args: [{
                table:'ts_ms_01',
                start: '2021-01-01 01:30:00',
                end: '2021-01-01 02:00:00',
                interval: '4 millisecond',
                expectedTimeGrain: 'ms'
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
                expectedTimeGrain: 'ms'
            }]}]
    },
    {
        title: '',
        subData: [{ args: [{
            table:'ts_seconds_01',
            start: '2021-01-01 12:30:00',
            end: '2021-01-01 13:30:00',
            interval: '2 seconds',
            expectedTimeGrain: 'second'
        }]}]
    },
    {
        title: '',
        subData: [{ args: [{
            table:'ts_minutes_01',
            start: '2021-01-01 01:04:04',
            end: '2021-01-01 09:04:04',
            interval: '47 minutes',
            expectedTimeGrain: 'minute'
        }]}]
    },
    {
        title: '',
        subData: [{ args: [{
            table:'ts_hours_01',
            start: '2021-01-01',
            end: '2022-01-01',
            interval: '2 hours',
            expectedTimeGrain: 'hour'
        }]}]
    },
    {
        title: '',
        subData: [{ args: [{
            table:'ts_days_01',
            start: '2021-01-01',
            end: '2025-01-01',
            interval: '3 days',
            expectedTimeGrain: 'day'
        }]}]
    },
    {
        title: '',
        subData: [{ args: [{
            table:'ts_weeks_01',
            start: '1900-01-01',
            end: '2000-01-01',
            interval: '7 day',
            expectedTimeGrain: 'week'
        }]}]
    },
    {
        title: '',
        subData: [{ args: [{
            table:'ts_weeks_02',
            start: '1900-01-01',
            end: '1900-03-01',
            interval: '7 day',
            expectedTimeGrain: 'week'
        }]}]
    },
    {
        title: '',
        subData: [{ args: [{
            title: "weekly, 100 years",
            table:'ts_weeks_03',
            start: '1900-01-01',
            end: '2000-01-01',
            interval: '7 day',
            expectedTimeGrain: 'week'
        }]}]
    },
    {
        title: '',
        subData: [{ args: [{
            table:'ts_years_01',
            start: '1900-01-01',
            end: '2000-01-01',
            interval: '1 year',
            expectedTimeGrain: 'year'
        }]}]
    },
]} as DataProviderData<[GeneratedTimeseriesTestCase]>

@FunctionalTestBase.Suite
export class StateSyncServiceSpec extends FunctionalTestBase  {
    protected databaseService: DatabaseService;

    public async setup(): Promise<void> {
        const config = new RootConfig({
            database: new DatabaseConfig({ databaseName: ":memory:" }),
            state: new StateConfig({ autoSync: true, syncInterval: 50 }),
            projectFolder: SYNC_TEST_FOLDER, profileWithUpdate: false,
        });
        await super.setup(config);
        const secondServerInstances = dataModelerServiceFactory(config);
        this.databaseService = secondServerInstances.dataModelerService.getDatabaseService();
        await this.databaseService.init();
    }
    public seriesGeneratedTimegrainData(): DataProviderData<[GeneratedTimeseriesTestCase]> {
        return subData;
    }

    @TestBase.Test("seriesGeneratedTimegrainData")
    public async shouldIdentifyTimegrain(args:GeneratedTimeseriesTestCase) {
        console.log(args.table)
        // generate the temporary table.
        // @ts-ignore
        await this.databaseService.databaseClient.execute(generateSeries(args.table, args.start, args.end, args.interval));
        const {timeGrain} = await this.databaseService.dispatch("estimateTimeGrain", [args.table, "ts"]);
        expect(args.expectedTimeGrain).toBe(timeGrain);
    }
}
