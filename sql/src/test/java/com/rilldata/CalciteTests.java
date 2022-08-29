package com.rilldata;

import com.rilldata.calcite.CalciteToolbox;
import com.rilldata.calcite.extensions.SqlCreateMetric;
import org.apache.calcite.sql.SqlNode;
import org.apache.calcite.sql.dialect.H2SqlDialect;
import org.apache.calcite.sql.parser.SqlParseException;
import org.apache.calcite.tools.Planner;
import org.apache.calcite.tools.ValidationException;
import org.apache.calcite.util.Litmus;
import org.junit.jupiter.api.Assertions;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.MethodSource;

import javax.sql.DataSource;
import java.sql.Connection;
import java.sql.SQLException;
import java.sql.Statement;
import java.util.Map;
import java.util.Optional;
import java.util.stream.Stream;

public class CalciteTests
{
  static CalciteToolbox calciteToolbox;

  @BeforeAll
  static void setUp() throws SQLException, ValidationException, SqlParseException
  {
    HsqlDbSchemaSupplier rootSchemaSupplier = new HsqlDbSchemaSupplier(Map.of("main", "PUBLIC"));
    calciteToolbox = new CalciteToolbox(
        rootSchemaSupplier,
        H2SqlDialect.DEFAULT,
        null
    );
    DataSource dataSource = rootSchemaSupplier.getDataSource();
    try (Connection conn = dataSource.getConnection(); Statement statement = conn.createStatement()) {
      statement.executeUpdate(
          "CREATE TABLE TEST("
              + "DIM1 INTEGER NOT NULL, "
              + "DIM2 VARCHAR(255), "
              + "DIM4 VARCHAR(255), "
              + "MET1 DOUBLE CHECK(MET1<10), "
              + "date DATE UNIQUE, "
              + "time TIMESTAMP)"
      );
    }
    String modelingQuery = """
        CREATE METRICS VIEW METRICS_VIEW
        DIMENSIONS
        DIM1, DIM2, ceil("MET1") AS DIM3
        MEASURES
        COUNT(DISTINCT DIM1) AS M_DIST,
        AVG(DISTINCT MET1) AS M_AVG
        FROM MAIN.TEST""";
    calciteToolbox.saveModel(modelingQuery);
  }

  @ParameterizedTest
  @MethodSource("testCreateMetricParams")
  public void testCreateMetric(String modelingQuery, int numDims, int numMeasures, Optional<String> parseExceptionMatch,
      Optional<String> validationExceptionMatch
  )
  {
    SqlCreateMetric sqlCreateMetric = null;
    try {
      sqlCreateMetric = calciteToolbox.parseModelingQuery(modelingQuery);
    } catch (SqlParseException e) {
      if (parseExceptionMatch.isEmpty() || !e.getMessage().contains(parseExceptionMatch.get())) {
        throw new RuntimeException(e);
      }
    }
    Assertions.assertEquals(numDims, sqlCreateMetric.dimensions.size());
    Assertions.assertEquals(numMeasures, sqlCreateMetric.measures.size());
    try {
      calciteToolbox.validateModelingQuery(sqlCreateMetric);
    } catch (SqlParseException e) {
      throw new RuntimeException(e);
    } catch (ValidationException e) {
      if (validationExceptionMatch.isEmpty() || !e.getMessage().contains(validationExceptionMatch.get())) {
        throw new RuntimeException(e);
      }
    }
  }

  @ParameterizedTest
  @MethodSource("testQueryExpansionParams")
  public void testQueryExpansion(String query, String expandedQuery, Optional<String> exceptionMessage)
      throws SqlParseException
  {
    try {
      String resultantQuery = calciteToolbox.getRunnableQuery(query);
      SqlNode actual = parseQuery(resultantQuery);
      SqlNode expected = parseQuery(expandedQuery);
      Assertions.assertTrue(exceptionMessage.isEmpty() && SqlNode.equalDeep(actual, expected, Litmus.IGNORE));
    } catch (RuntimeException | ValidationException e) {
      Assertions.assertTrue(exceptionMessage.isPresent() && e.getMessage().contains(exceptionMessage.get()));
    }
  }

  // used to get AST for comparison instead of comparing Strings
  public SqlNode parseQuery(String sql) throws SqlParseException
  {
    Planner planner = calciteToolbox.getPlanner();
    SqlNode sqlNode = planner.parse(sql);
    planner.close();
    return sqlNode;
  }

  public static Stream<Arguments> testQueryExpansionParams()
  {
    return Stream.of(
        Arguments.of(
            "SELECT DIM1, M_DIST FROM METRICS_VIEW",
            "SELECT DIM1, COUNT(DISTINCT DIM1) AS M_DIST FROM MAIN.TEST GROUP BY 1",
            Optional.empty()
        ),
        Arguments.of(
            "SELECT DIM1, DIM3, M_DIST, M_AVG FROM METRICS_VIEW",
            "SELECT DIM1, ceil(\"MET1\") AS DIM3, COUNT(DISTINCT DIM1) AS M_DIST, AVG(DISTINCT MET1) AS M_AVG "
                + "FROM MAIN.TEST GROUP BY 1,2",
            Optional.empty()
        ),
        Arguments.of(
            "SELECT DIM1, DIM3 AS D3, M_DIST FROM METRICS_VIEW", // using alias for specified measure alias
            "SELECT DIM1, ceil(\"MET1\") AS D3, COUNT(DISTINCT DIM1) AS M_DIST FROM MAIN.TEST GROUP BY 1,2",
            Optional.empty()
        ),
        Arguments.of(
            "SELECT DIM1 AS D1, M_DIST AS MD FROM METRICS_VIEW", // using aliases for dimensions
            "SELECT DIM1 AS D1, COUNT(DISTINCT DIM1) AS MD FROM MAIN.TEST GROUP BY 1",
            Optional.empty()
        ),
        Arguments.of(
            "SELECT DIM1, DIM3, DIM2 FROM METRICS_VIEW", // no measures, resultant query is simple select not group by
            "SELECT DIM1, ceil(\"MET1\") AS DIM3, DIM2 FROM MAIN.TEST",
            Optional.empty()
        ),
        Arguments.of(
            "SELECT M_DIST, M_AVG FROM METRICS_VIEW", // no dimensions, resultant query is simple select not group by
            "SELECT COUNT(DISTINCT DIM1) AS M_DIST, AVG(DISTINCT MET1) AS M_AVG FROM MAIN.TEST",
            Optional.empty()
        ),
        Arguments.of(
            "SELECT DIM1, M_DIST FROM METRICS_VIEW LIMIT 5", // limit works
            "SELECT DIM1, COUNT(DISTINCT DIM1) AS M_DIST FROM MAIN.TEST GROUP BY 1 LIMIT 5",
            Optional.empty()
        ),
        Arguments.of(
            "SELECT DIM1, M_DIST FROM METRICS_VIEW ORDER BY DIM1 LIMIT 5", // order by works
            "SELECT DIM1, COUNT(DISTINCT DIM1) AS M_DIST FROM MAIN.TEST GROUP BY 1 ORDER BY DIM1 LIMIT 5",
            Optional.empty()
        ),
        Arguments.of(
            "SELECT DIM1, M_DIST FROM METRICS_VIEW ORDER BY M_DIST LIMIT 5", // order by a metric works
            "SELECT DIM1, COUNT(DISTINCT DIM1) AS M_DIST FROM MAIN.TEST GROUP BY 1 ORDER BY M_DIST LIMIT 5",
            Optional.empty()
        ),
        Arguments.of(
            "SELECT DIM1, M_DIST FROM METRICS_VIEW ORDER BY DIM1, M_DIST LIMIT 5", // order by a multiple cols works
            "SELECT DIM1, COUNT(DISTINCT DIM1) AS M_DIST FROM MAIN.TEST GROUP BY 1 ORDER BY DIM1, M_DIST LIMIT 5",
            Optional.empty()
        ),
        Arguments.of(
            "SELECT DIM1 AS D1, M_DIST FROM METRICS_VIEW ORDER BY D1 LIMIT 5", // order by with alias
            "SELECT DIM1 AS D1, COUNT(DISTINCT DIM1) AS M_DIST FROM MAIN.TEST GROUP BY 1 ORDER BY D1 LIMIT 5",
            Optional.empty()
        ),
        Arguments.of(
            "SELECT DIM1 AS D1, M_DIST FROM METRICS_VIEW ORDER BY D1 DESC LIMIT 5", // order by desc works
            "SELECT DIM1 AS D1, COUNT(DISTINCT DIM1) AS M_DIST FROM MAIN.TEST GROUP BY 1 ORDER BY D1 DESC LIMIT 5",
            Optional.empty()
        ),
        Arguments.of(
            "SELECT DIM1 AS D1, M_DIST FROM METRICS_VIEW WHERE DIM1='something' ORDER BY D1 DESC LIMIT 5", // where clause works
            "SELECT DIM1 AS D1, COUNT(DISTINCT DIM1) AS M_DIST FROM MAIN.TEST WHERE DIM1='something' GROUP BY 1 ORDER BY D1 DESC LIMIT 5",
            Optional.empty()
        ),
        Arguments.of(
            "SELECT DIM1 AS D1, M_DIST FROM METRICS_VIEW WHERE DIM4='something'",
            // where clause works for any column present in base table but may not be in METRICS VIEW, should we disallow ?
            "SELECT DIM1 AS D1, COUNT(DISTINCT DIM1) AS M_DIST FROM MAIN.TEST WHERE DIM4='something' GROUP BY 1",
            Optional.empty()
        ),
        Arguments.of(
            "SELECT DIM1 AS D1, M_DIST FROM METRICS_VIEW WHERE NOT_PRESENT='something'",
            "",
            Optional.of("Column 'NOT_PRESENT' not found in any table")
        ),
        Arguments.of(
            "SELECT DIM1 FROM (SELECT DIM1, M_DIST FROM METRICS_VIEW)", // inner query works
            "SELECT DIM1 FROM (SELECT DIM1, COUNT(DISTINCT DIM1) AS M_DIST FROM MAIN.TEST GROUP BY 1)",
            Optional.empty()
        ),
        Arguments.of(
            "SELECT D, M FROM (SELECT DIM1 AS D, M_DIST AS M FROM METRICS_VIEW)", // alias with inner query
            "SELECT D, M FROM (SELECT DIM1 AS D, COUNT(DISTINCT DIM1) AS M FROM MAIN.TEST GROUP BY 1)",
            Optional.empty()
        ),
        Arguments.of(
            "SELECT D, M FROM (SELECT DIM1 AS D, M_DIST AS M FROM METRICS_VIEW WHERE DIM2='something')",
            // where clause in inner query works
            "SELECT D, M FROM (SELECT DIM1 AS D, COUNT(DISTINCT DIM1) AS M FROM MAIN.TEST WHERE DIM2='something' GROUP BY 1)",
            Optional.empty()
        ),
        Arguments.of(
            "SELECT M_DIST, DIM1, M_AVG FROM METRICS_VIEW", // rearranges dimensions to start of list
            "SELECT DIM1, COUNT(DISTINCT DIM1) AS M_DIST, AVG(DISTINCT MET1) AS M_AVG FROM MAIN.TEST GROUP BY 1",
            Optional.empty()
        ),
        Arguments.of(
            "SELECT M_DIST, DIM2, M_AVG, DIM1 FROM METRICS_VIEW",
            // rearranges dimensions to start of list in same order as original query
            "SELECT DIM2, DIM1, COUNT(DISTINCT DIM1) AS M_DIST, AVG(DISTINCT MET1) AS M_AVG FROM MAIN.TEST GROUP BY 1,2",
            Optional.empty()
        ),
        Arguments.of(
            "SELECT dim1, M FROM (SELECT dim1, M_DIST AS M FROM METRICS_VIEW)", // not case-sensitive
            "SELECT dim1, M FROM (SELECT dim1, COUNT(DISTINCT DIM1) AS M FROM MAIN.TEST GROUP BY 1)",
            Optional.empty()
        ),
        Arguments.of(
            "SELECT * FROM METRICS_VIEW", // star works
            "SELECT DIM1, DIM2, ceil(\"MET1\") AS DIM3, COUNT(DISTINCT DIM1) AS M_DIST, AVG(DISTINCT MET1) AS M_AVG "
                + "FROM MAIN.TEST GROUP BY 1,2,3",
            Optional.empty()
        ),
        Arguments.of(
            "SELECT *, DIM1 FROM METRICS_VIEW", // cannot have another col with star
            // this is something which DuckDB supports, we may need to look into supporting it in the future.
            "",
            Optional.of("Cannot specify columns along with *")
        ),
        Arguments.of(
            "SELECT DIM1, DIM3, M_DIST, DIM4 FROM METRICS_VIEW", // cannot use column which was not present in metrics view
            "",
            Optional.of("Column [DIM4] not present in metrics view [METRICS_VIEW]")
        ),
        Arguments.of(
            "SELECT T.DIM1, T.M_DIST FROM METRICS_VIEW T", // Model alias not supported as of now, look into it later
            "",
            Optional.of("Column [T.DIM1] not present in metrics view [METRICS_VIEW]")
        ),
        Arguments.of(
            """
                WITH\s
                CTE1 AS (\s
                SELECT DIM1, M_DIST FROM METRICS_VIEW
                )\s
                SELECT * FROM CTE1""",
            """
                WITH\s
                CTE1 AS (\s
                SELECT DIM1, COUNT(DISTINCT DIM1) AS M_DIST FROM MAIN.TEST GROUP BY 1
                )\s
                SELECT * FROM CTE1""",
            Optional.empty()
        )
    );
  }

  private static Stream<Arguments> testCreateMetricParams()
  {
    return Stream.of(
        Arguments.of("""
                CREATE METRICS VIEW METRICS_VIEW
                DIMENSIONS
                DIM1, ceil("MET1") AS DIM3, DIM2
                MEASURES
                COUNT(DISTINCT DIM1) AS M_DIST,
                AVG(DISTINCT MET1) AS M_AVG
                FROM MAIN.TEST""",
            3,
            2,
            Optional.empty(),
            Optional.empty()
        ),
        Arguments.of("""
                CREATE METRICS VIEW Test2
                DIMENSIONS
                DIM1, MET1
                MEASURES
                AVG(MET1) AS M_AVG
                FROM MAIN.TEST""",
            2,
            1,
            Optional.empty(),
            Optional.empty()
        ),
        Arguments.of("""
                CREATE METRICS VIEW Test3
                DIMENSIONS
                "date", current_timestamp AS CURR
                MEASURES
                last_day("time")
                FROM MAIN.TEST""",
            2,
            1,
            Optional.empty(),
            Optional.of("Expression 'time' is not being grouped")
        ),
        // currently not all duckdb functions are supported, this should ideally pass
        Arguments.of("""
                CREATE METRICS VIEW Test4
                DIMENSIONS
                DATE_TRUNC('year', "date"), current_timestamp AS CURR
                MEASURES
                COUNT(*) AS C_STAR
                FROM MAIN.TEST""",
            2,
            1,
            Optional.empty(),
            Optional.of("No match found for function signature DATE_TRUNC(<CHARACTER>, <DATE>)")
        ),
        Arguments.of("""
                CREATE METRICS VIEW Test5
                DIMENSIONS
                FAKE
                MEASURES
                COUNT(*) AS C_STAR
                FROM MAIN.TEST""",
            1,
            1,
            Optional.empty(),
            Optional.of("Column 'FAKE' not found in any table")
        )
    );
  }
}
