package com.rilldata;

import com.google.protobuf.InvalidProtocolBufferException;
import com.rilldata.calcite.CalciteToolbox;
import com.rilldata.calcite.dialects.Dialects;
import com.rilldata.calcite.models.Artifact;
import com.rilldata.calcite.models.ArtifactStore;
import com.rilldata.calcite.models.ArtifactType;
import com.rilldata.calcite.models.SqlCreateMetricsView;
import com.rilldata.calcite.models.SqlCreateSource;
import com.rilldata.protobuf.generated.SqlNodeProto;
import org.apache.calcite.sql.SqlNode;
import org.apache.calcite.sql.parser.SqlParseException;
import org.apache.calcite.tools.Planner;
import org.apache.calcite.tools.ValidationException;
import org.apache.calcite.util.Litmus;
import org.junit.jupiter.api.Assertions;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.MethodSource;

import javax.sql.DataSource;
import java.sql.Connection;
import java.sql.SQLException;
import java.sql.Statement;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.stream.Stream;

public class CalciteTests
{
  static CalciteToolbox calciteToolbox;

  @BeforeAll
  static void setUp() throws SQLException
  {
    HsqlDbSchemaSupplier rootSchemaSupplier = new HsqlDbSchemaSupplier(Map.of("main", "PUBLIC"));
    calciteToolbox = new CalciteToolbox(rootSchemaSupplier, new ArtifactStore(
        List.of(new Artifact(
            "METRICS_VIEW",
            ArtifactType.METRICS_VIEW,
            """
                CREATE METRICS VIEW METRICS_VIEW
                DIMENSIONS
                DIM1, DIM2, ceil("MET1") AS DIM3
                MEASURES
                COUNT(DISTINCT DIM1) AS M_DIST,
                AVG(DISTINCT MET1) AS M_AVG
                FROM MAIN.TEST"""
        ))
    ));
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
      statement.executeUpdate("create table heroes (id smallint, power tinyint, name varchar(255))");
    }
  }

  @ParameterizedTest
  @MethodSource("testCreateMetricsViewParams")
  public void testCreateMetricsView(String modelingQuery, int numDims, int numMeasures,
      Optional<String> parseExceptionMatch, Optional<String> validationExceptionMatch
  )
  {
    SqlCreateMetricsView sqlCreateMetricsView;
    byte[] ast;
    try {
      ast = calciteToolbox.getAST(modelingQuery, false);
      parseExceptionMatch.ifPresent(s -> System.out.println("Expected following exception : " + s));
      Assertions.assertTrue(parseExceptionMatch.isEmpty());
      validationExceptionMatch.ifPresent(s -> System.out.println("Expected following exception : " + s));
      Assertions.assertTrue(validationExceptionMatch.isEmpty());
    } catch (SqlParseException e) {
      if (parseExceptionMatch.isEmpty() || !e.getMessage().contains(parseExceptionMatch.get())) {
        e.printStackTrace();
      }
      Assertions.assertTrue(parseExceptionMatch.isPresent() && e.getMessage().contains(parseExceptionMatch.get()));
      return; // found parse exception - test done - return now
    } catch (ValidationException e) {
      if (validationExceptionMatch.isEmpty() || !e.getMessage().contains(validationExceptionMatch.get())) {
        e.printStackTrace();
      }
      Assertions.assertTrue(
          validationExceptionMatch.isPresent() && e.getMessage().contains(validationExceptionMatch.get()));
      return; // found validation exception - test done - return now
    }
    try {
      sqlCreateMetricsView = (SqlCreateMetricsView) calciteToolbox.parseValidatedSql(modelingQuery);
    } catch (SqlParseException e) {
      throw new RuntimeException(e);
    }
    Assertions.assertEquals(numDims, sqlCreateMetricsView.dimensions.size());
    Assertions.assertEquals(numMeasures, sqlCreateMetricsView.measures.size());
    try {
      SqlNodeProto sqlNodeProto = SqlNodeProto.parseFrom(ast);
      Assertions.assertTrue(sqlNodeProto.toString().length() > 0);
    } catch (InvalidProtocolBufferException e) {
      throw new RuntimeException(e);
    }
  }

  @ParameterizedTest
  @MethodSource("testCreateSourceParams")
  public void testCreateSource(String createSourceQuery, Optional<String> parseExceptionMatch,
      Optional<String> validationExceptionMatch
  )
  {
    byte[] ast;
    try {
      ast = calciteToolbox.getAST(createSourceQuery, false);
      parseExceptionMatch.ifPresent(s -> System.out.println("Expected following exception : " + s));
      Assertions.assertTrue(parseExceptionMatch.isEmpty());
      validationExceptionMatch.ifPresent(s -> System.out.println("Expected following exception : " + s));
      Assertions.assertTrue(validationExceptionMatch.isEmpty());
      SqlCreateSource sqlCreateSource = (SqlCreateSource) calciteToolbox.parseValidatedSql(createSourceQuery);
      Assertions.assertNotNull(sqlCreateSource);
    } catch (SqlParseException e) {
      if (parseExceptionMatch.isEmpty() || !e.getMessage().contains(parseExceptionMatch.get())) {
        e.printStackTrace();
      }
      Assertions.assertTrue(parseExceptionMatch.isPresent() && e.getMessage().contains(parseExceptionMatch.get()));
      return; // found parse exception - test done - return now
    } catch (ValidationException e) {
      if (validationExceptionMatch.isEmpty() || !e.getMessage().contains(validationExceptionMatch.get())) {
        e.printStackTrace();
      }
      Assertions.assertTrue(
          validationExceptionMatch.isPresent() && e.getMessage().contains(validationExceptionMatch.get()));
      return; // found validation exception - test done - return now
    }
    try {
      SqlNodeProto sqlNodeProto = SqlNodeProto.parseFrom(ast);
      Assertions.assertTrue(sqlNodeProto.toString().length() > 0);
    } catch (InvalidProtocolBufferException e) {
      throw new RuntimeException(e);
    }
  }

  private static Stream<Arguments> testCreateSourceParams()
  {
    return Stream.of(
        Arguments.of("""
                CREATE SOURCE clicks_raw
                WITH (
                    'connector' = 's3',
                    'prefix' = 's3://my_bucket/a.csv', // comments are ignored
                    'FORMAT' = 'CSV'
                )""",
            Optional.empty(),
            Optional.empty()
        ),
        Arguments.of("""
                CREATE SOURCE clicks_raw
                WITH (
                    'connector' = 's3',
                    'prefix' = 's3://my_bucket/a.csv', // comments are ignored
                    'FORMAT' = 'CSV', // extra comma at the is ignored
                )""",
            Optional.empty(),
            Optional.empty()
        ),
        Arguments.of("""
                CREATE SOURCE clicks_raw
                WITH (
                    'connector' = 's3',
                    'prefix' = 's3://my_bucket/a.csv', // comments are ignored
                    'FORMAT' = 'CSV',, // extra commas
                )""",
            Optional.of("Encountered \" \",\" \", \"\""),
            Optional.empty()
        ),
        Arguments.of("""
                CREATE SOURCE clicks_raw
                WITH ( 'connector' = 's3', 'prefix' = 's3://my_bucket/a.csv', 'FORMAT' = 'CSV',)""",
            Optional.empty(),
            Optional.empty()
        ),
        // Paranthesis are optional
        Arguments.of("""
                CREATE SOURCE clicks_raw
                WITH 'connector' = 's3', 'prefix' = 's3://my_bucket/a.csv', 'FORMAT' = 'CSV'""",
            Optional.empty(),
            Optional.empty()
        ),
        Arguments.of("""
                CREATE SOURCE clicks_raw
                WITH
                'connector' = 's3',
                'prefix' = 's3://my_bucket/a.csv',
                'FORMAT' = 'CSV'""",
            Optional.empty(),
            Optional.empty()
        ),
        Arguments.of("""
                CREATE SOURCE clicks_raw
                WITH 'connector' = 's3', 'prefix' = 's3://my_bucket/a.csv', 'FORMAT' = 'CSV',""",
            Optional.empty(),
            Optional.empty()
        ),
        Arguments.of("""
                CREATE SOURCE clicks_raw
                WITH ('connector' = 's3', 'prefix' = 's3://my_bucket/a.csv', 'FORMAT' = 'CSV',""",
            Optional.of("""
                Encountered "<EOF>" at line 2, column 78.
                Was expecting:
                    ")" ..."""),
            Optional.empty()
        ),
        Arguments.of("""
                CREATE SOURCE clicks_raw
                WITH 'connector' = 's3', 'prefix' = 's3://my_bucket/a.csv', 'FORMAT' = 'CSV',)""",
            Optional.of("""
                Encountered " ")" ") "" at line 2, column 78.
                Was expecting:
                    <EOF>"""),
            Optional.empty()
        ),
        Arguments.of("""
                CREATE SOURCE clicks_raw
                WITH (
                    'connector' = 's3',, // extra comma
                    'prefix' = 's3://my_bucket/a.csv', // comments are ignored
                    'FORMAT' = 'CSV', // extra comma at the is ignored
                )""",
            Optional.of("Encountered \" \",\" \", \"\""),
            Optional.empty()
        ),
        Arguments.of("""
                CREATE SOURCE clicks_raw
                WITH (
                    'connector' = 's3' // missing comma
                    'prefix' = 's3://my_bucket/a.csv',
                    'FORMAT' = 'CSV',
                )""",
            Optional.of("Encountered \" \"=\" \"= \"\""),
            Optional.empty()
        ),
        // This should not fail with parse exception, parsing should pass
        Arguments.of("""
                CREATE SOURCE clicks_raw
                WITH (
                    'connector' = 's3'
                )""",
            Optional.empty(),
            Optional.of("Required property [prefix] not present or blank for s3 connector")
        ),
        // This should not fail with parse exception, pasring should pass
        Arguments.of("""
                CREATE SOURCE clicks_raw
                WITH (
                    'connector' = 's3',
                )""",
            Optional.empty(),
            Optional.of("Required property [prefix] not present or blank for s3 connector")
        ),
        Arguments.of("""
                CREATE SOURCE clicks_raw
                WITH (
                    -- comments are ignored
                    'connector' = 's3',
                    'prefix' = 's3://my_bucket/a.csv', -- comments are ignored
                    'FORMAT' = 'CSV'
                )""",
            Optional.empty(),
            Optional.empty()
        ),
        Arguments.of("""
                CREATE SOURCE clicks_raw
                WITH (
                    -- comments are ignored
                    'connector' = 's3',
                    'prefix' = 's3://my_bucket/a.csv',
                    'FORMAT' = 'CSV', // extra comma
                )""",
            Optional.empty(),
            Optional.empty()
        ),
        Arguments.of("""
                CREATE SOURCE clicks_raw
                WITH (
                    -- comments are ignored
                    'connector' = 's3',
                    'prefix' = 's3://my_bucket/a.csv' -- comments are ignored
                )""",
            Optional.empty(),
            Optional.of("Required property [format] not present or blank for s3 connector")
        ),
        Arguments.of("""
                CREATE SOURCE clicks_raw
                WITH (
                    -- comments are ignored
                    'connector' = 's3',
                    'prefix' = 's3://my_bucket/a.csv',
                    'format' = '' -- empty
                )""",
            Optional.empty(),
            Optional.of("Required property [format] not present or blank for s3 connector")
        ),
        Arguments.of("""
                CREATE SOURCE clicks_raw
                WITH (
                    'connector' = 's3',
                    'prefix' = 's3://my_bucket/a.csv',
                    'format' = 'myformat'
                )""",
            Optional.empty(),
            Optional.of("Format [myformat] not supported, supported formats are")
        ),
        Arguments.of("""
                CREATE SOURCE clicks_raw
                WITH (
                    -- comments are ignored
                    'prefix' = 's3://my_bucket/a.csv' -- comments are ignored
                )""",
            Optional.empty(),
            Optional.of("Required property [connector] not found for source [CLICKS_RAW]")
        ),
        // TODO this probably should pass
        Arguments.of("""
                CREATE SOURCE clicks_raw
                WITH (
                    'connector' = 's3',
                    'format' = 'csv',
                    'prefix' = 's3://my_bucket/*.csv',
                    'aws.access.key' = env('S3_ACCESS_KEY'),
                    'aws.secret.key' = env('S3_SECRET_KEY')
                )""",
            Optional.of("Encountered \" <IDENTIFIER> \"env \"\""),
            Optional.empty()
        ),
        Arguments.of("""
                CREATE SOURCE clicks_raw
                WITH (
                    'connector' = 's4',
                    'format' = 'csv',
                    'prefix' = 's3://my_bucket/*.csv'
                )""",
            Optional.empty(),
            Optional.of("No connector of type [s4] found for source [CLICKS_RAW]")
        ),
        // without quotes property key is parsed as an identifier
        Arguments.of("""
                CREATE SOURCE clicks_raw
                WITH (
                    connector = 's3',
                    'format' = 'csv',
                    'prefix' = 's3://my_bucket/*.csv'
                )""",
            Optional.empty(),
            Optional.empty()
        ),
        Arguments.of("""
                CREATE SOURCE clicks_raw
                WITH (
                    connector = 's3',
                    'format' = 'csv',
                    prefix = 's3://my_bucket/*.csv',
                )""",
            Optional.empty(),
            Optional.empty()
        ),
        Arguments.of("""
                CREATE SOURCE clicks_raw
                WITH (
                    'connector' = 's3',
                    format = 'csv',
                    prefix = 's3://my_bucket/*.csv',
                )""",
            Optional.empty(),
            Optional.empty()
        ),
        Arguments.of("""
                CREATE SOURCE clicks_raw
                WITH (
                    connector = 's3',
                    format = 'csv',
                    prefix = 's3://my_bucket/*.csv',
                )""",
            Optional.empty(),
            Optional.empty()
        ),
        // property needs to be a simple identifier
        Arguments.of("""
                CREATE SOURCE clicks_raw
                WITH (
                    connector.source = 's3',
                    format = 'csv',
                    prefix = 's3://my_bucket/*.csv',
                )""",
            Optional.of("""
                Encountered " "." ". "" at line 3, column 14.
                Was expecting:
                    "=" ..."""),
            Optional.empty()
        ),
        // parsing will pass but validation will fail because required property "connector" is missing
        Arguments.of("""
                CREATE SOURCE clicks_raw
                WITH (
                    connector$source = 's3',
                    format = 'csv',
                    prefix = 's3://my_bucket/*.csv',
                )""",
            Optional.empty(),
            Optional.of("Required property [connector] not found for source [CLICKS_RAW]")
        ),
        // property value needs to be quoted
        Arguments.of("""
                CREATE SOURCE clicks_raw
                WITH (
                    connector = 's3',
                    format = 'csv',
                    prefix = 's3://my_bucket/*.csv',
                    key123 = value123
                )""",
            Optional.of("Encountered \" <IDENTIFIER> \"value123 \"\""),
            Optional.empty()
        ),
        Arguments.of("""
                CREATE SOURCE clicks_raw
                WITH (
                    connector = 's3',
                    format = 'csv',
                    prefix = 's3://my_bucket/*.csv',
                    key123 = 'value123'
                )""",
            Optional.empty(),
            Optional.empty()
        ),
        Arguments.of("""
                CREATE SOURCE clicks_raw
                WITH (
                    connector = 's3',
                    format = 'csv',
                    prefix = 's3://my_bucket/*.csv',
                    key = '123'
                )""",
            Optional.empty(),
            Optional.empty()
        ),
        Arguments.of("""
                CREATE SOURCE clicks_raw
                WITH (
                    connector = 's3',
                    format = 'csv',
                    prefix = 's3://my_bucket/*.csv',
                    '123' = '123'
                )""",
            Optional.empty(),
            Optional.empty()
        ),
        // if the proper key contains special characters, it needs to be quoted
        Arguments.of("""
                CREATE SOURCE clicks_raw
                WITH (
                    connector = 's3',
                    format = 'csv',
                    prefix = 's3://my_bucket/*.csv',
                    a123@ = '123'
                )""",
            Optional.of("Lexical error at line 6, column 9.  Encountered: \"@\""),
            Optional.empty()
        ),
        // if the proper key contains special characters, it needs to be quoted
        Arguments.of("""
                CREATE SOURCE clicks_raw
                WITH (
                    connector = 's3',
                    format = 'csv',
                    prefix = 's3://my_bucket/*.csv',
                    'a123@' = '123'
                )""",
            Optional.empty(),
            Optional.empty()
        ),
        // if the proper key contains only numbers, it needs to be quoted
        Arguments.of("""
                CREATE SOURCE clicks_raw
                WITH (
                    connector = 's3',
                    format = 'csv',
                    prefix = 's3://my_bucket/*.csv',
                    123 = '123'
                )""",
            Optional.of("Encountered \" <UNSIGNED_INTEGER_LITERAL> \"123 \"\""),
            Optional.empty()
        ),
        // if the proper key contains only numbers, it needs to be quoted
        Arguments.of("""
                CREATE SOURCE clicks_raw
                WITH (
                    connector = 's3',
                    format = 'csv',
                    prefix = 's3://my_bucket/*.csv',
                    '123' = '123'
                )""",
            Optional.empty(),
            Optional.empty()
        ),
        Arguments.of("""
                CREATE SOURCE clicks_raw
                WITH (
                    connector = 's3',
                    format = 'csv',
                    prefix = 's3://my_bucket/*.csv',
                    a@123 = '123'
                )""",
            Optional.of("Lexical error at line 6, column 6.  Encountered: \"@\""),
            Optional.empty()
        )
    );
  }

  @Test
  public void testSimpleQuery() throws SqlParseException, ValidationException
  {
    String query = "select 1 as foo, "
        + "'hello' as bar, h1.id, h1.\"power\", h2.name "
        + "from main.heroes h1 join main.heroes h2 on h1.id = h2.id";
    for (Dialects dialect : Dialects.values()) {
      calciteToolbox.getRunnableQuery(query, dialect.getSqlDialect());
    }
  }

  @ParameterizedTest
  @MethodSource("testQueryExpansionParams")
  public void testQueryExpansion(String query, String expandedQuery, Optional<String> exceptionMessage)
      throws SqlParseException
  {
    try {
      for (Dialects dialect : Dialects.values()) {
        String resultantQuery = calciteToolbox.getRunnableQuery(query, dialect.getSqlDialect());
        String expectedQuery = calciteToolbox.getRunnableQuery(expandedQuery, dialect.getSqlDialect());
        SqlNode actual = parseQuery(resultantQuery);
        SqlNode expected = parseQuery(expectedQuery);
        exceptionMessage.ifPresent(s -> System.out.println("Expected following exception : " + s));
        Assertions.assertTrue(exceptionMessage.isEmpty() && SqlNode.equalDeep(actual, expected, Litmus.IGNORE));

        byte[] ast = calciteToolbox.getAST(actual);
        try {
          SqlNodeProto sqlNodeProto = SqlNodeProto.parseFrom(ast);
          Assertions.assertTrue(sqlNodeProto.toString().length() > 0);
        } catch (InvalidProtocolBufferException e) {
          throw new RuntimeException(e);
        }
      }
    } catch (RuntimeException | ValidationException e) {
      if (exceptionMessage.isEmpty() || !e.getMessage().contains(exceptionMessage.get())) {
        e.printStackTrace();
      }
      Assertions.assertTrue(exceptionMessage.isPresent() && e.getMessage().contains(exceptionMessage.get()));
    }
  }

  @ParameterizedTest
  @MethodSource("testOperatorsParams")
  public void testOperators(String query, String expectedDuckDBQuery, String expectedDruidQuery,
      Optional<String> exceptionMessage
  ) throws SqlParseException
  {
    try {
      for (Dialects dialect : Dialects.values()) {
        String resultantQuery = calciteToolbox.getRunnableQuery(query, dialect.getSqlDialect());
        SqlNode actual = parseQuery(resultantQuery);
        SqlNode expected;
        if (dialect.equals(Dialects.DUCKDB)) {
          expected = parseQuery(expectedDuckDBQuery);
        } else {
          expected = parseQuery(expectedDruidQuery);
        }
        exceptionMessage.ifPresent(s -> System.out.println("Expected following exception : " + s));
        Assertions.assertTrue(exceptionMessage.isEmpty() && SqlNode.equalDeep(actual, expected, Litmus.IGNORE));

        byte[] ast = calciteToolbox.getAST(actual);
        try {
          SqlNodeProto sqlNodeProto = SqlNodeProto.parseFrom(ast);
          Assertions.assertTrue(sqlNodeProto.toString().length() > 0);
        } catch (InvalidProtocolBufferException e) {
          throw new RuntimeException(e);
        }
      }
    } catch (RuntimeException | ValidationException e) {
      if (exceptionMessage.isEmpty() || !e.getMessage().contains(exceptionMessage.get())) {
        e.printStackTrace();
      }
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
            "SELECT DIM1 AS D1, M_DIST FROM METRICS_VIEW WHERE DIM1='something' ORDER BY D1 DESC LIMIT 5",
            // where clause works
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
            "SELECT DIM1, DIM3, M_DIST, DIM4 FROM METRICS_VIEW",
            // cannot use column which was not present in metrics view
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

  public static Stream<Arguments> testOperatorsParams()
  {
    return Stream.of(
        Arguments.of(
            "SELECT GREATEST(1,2)",
            "SELECT GREATEST(1,2)",
            "SELECT GREATEST(1,2)",
            Optional.empty()
        ),
        Arguments.of(
            "SELECT GREATEST(1,'a')",
            "SELECT GREATEST(1,'a')",
            "SELECT GREATEST(1,'a')",
            Optional.empty()
        ),
        Arguments.of(
            "SELECT GREATEST('ABC','DEF')",
            "SELECT GREATEST('ABC', 'DEF')",
            "SELECT GREATEST('ABC', 'DEF')",
            Optional.empty()
        ),
        Arguments.of(
            "SELECT LEAST('ABC','DEF')",
            "SELECT LEAST('ABC', 'DEF')",
            "SELECT LEAST('ABC', 'DEF')",
            Optional.empty()
        ),
        Arguments.of(
            "SELECT LEAST(1,'a', 100)",
            "SELECT LEAST(1,'a', 100)",
            "SELECT LEAST(1,'a', 100)",
            Optional.empty()
        ),
        Arguments.of(
            "SELECT LOG(1)",
            "SELECT LOG(1)",
            "SELECT LOG(1)",
            Optional.empty()
        ),
        Arguments.of(
            "SELECT LOG2(1)",
            "SELECT LOG2(1)",
            "SELECT LOG2(1)",
            Optional.empty()
        ),
        Arguments.of(
            "SELECT LOG2('ABC')",
            "",
            "",
            Optional.of(
                "Cannot apply 'LOG2' to arguments of type 'LOG2(<CHAR(3)>)'. Supported form(s): 'LOG2(<NUMERIC>)")
        ),
        Arguments.of(
            "SELECT XOR(1,2)",
            "SELECT XOR(1,2)",
            "SELECT BITWISE_XOR(1,2)",
            Optional.empty()
        ),
        Arguments.of(
            "SELECT DATE_TRUNC('year', TIMESTAMP '2022-08-01')",
            "SELECT DATE_TRUNC('year', TIMESTAMP '2022-08-01')",
            "SELECT DATE_TRUNC('year', TIMESTAMP '2022-08-01')",
            Optional.empty()
        ),
        Arguments.of(
            "SELECT DATE_TRUNC('year', '2022-09-01')", // sql literal
            "",
            "",
            Optional.of(
                "Cannot apply 'DATE_TRUNC' to arguments of type 'DATE_TRUNC(<CHAR(4)>, <CHAR(10)>)'. Supported form(s): 'DATE_TRUNC(<CHARACTER>, <TIMESTAMP>)'")
        ),
        // ideally DATE can be implicitly cast to TIMESTAMP so that this one passes
        Arguments.of(
            "SELECT DATE_TRUNC('year', DATE '2022-08-01')", // druid only supports timestamp in date_trunc
            "",
            "",
            Optional.of(
                "Cannot apply 'DATE_TRUNC' to arguments of type 'DATE_TRUNC(<CHAR(4)>, <DATE>)'. Supported form(s): 'DATE_TRUNC(<CHARACTER>, <TIMESTAMP>)'")
        ),
        Arguments.of(
            "SELECT DATE_TRUNC('year', CAST(DATE '2022-08-01' AS TIMESTAMP))", // date column
            "SELECT DATE_TRUNC('year', CAST(DATE '2022-08-01' AS TIMESTAMP))",
            "SELECT DATE_TRUNC('year', CAST(DATE '2022-08-01' AS TIMESTAMP))",
            Optional.empty()
        ),
        Arguments.of(
            "SELECT XOR(2)",
            "",
            "",
            Optional.of("Invalid number of arguments to function 'XOR'")
        ),
        Arguments.of(
            "SELECT XOR('ABC', 'DEF')",
            "",
            "",
            Optional.of(
                "Cannot apply 'XOR' to arguments of type 'XOR(<CHAR(3)>, <CHAR(3)>)'. Supported form(s): 'XOR(<INTEGER>, <INTEGER>)'")
        )
    );
  }

  private static Stream<Arguments> testCreateMetricsViewParams()
  {
    return Stream.of(
        Arguments.of("""
                CREATE METRICS VIEW METRICS_VIEW
                DIMENSIONS // comments are ignored
                DIM1, ceil("MET1") AS DIM3, DIM2
                -- comments are ignored
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
        Arguments.of("""
                CREATE METRICS VIEW Test4
                DIMENSIONS
                DATE_TRUNC('year', CAST("date" AS TIMESTAMP)), current_timestamp AS CURR
                MEASURES
                COUNT(*) AS C_STAR
                FROM MAIN.TEST""",
            2,
            1,
            Optional.of("Please provide an alias for `DATE_TRUNC`('year', CAST(`date` AS TIMESTAMP))"),
            Optional.empty()
        ),
        Arguments.of("""
                CREATE METRICS VIEW Test4
                DIMENSIONS
                DATE_TRUNC('year', CAST("date" AS TIMESTAMP)) AS "the-year", current_timestamp AS CURR
                MEASURES
                COUNT(*) AS C_STAR
                FROM MAIN.TEST""",
            2,
            1,
            Optional.empty(),
            Optional.empty()
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
