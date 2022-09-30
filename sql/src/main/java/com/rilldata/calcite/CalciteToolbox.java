package com.rilldata.calcite;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.DeserializationFeature;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.rilldata.calcite.extensions.SqlCreateMetric;
import com.rilldata.calcite.generated.RillSqlParserImpl;
import com.rilldata.calcite.models.Artifact;
import com.rilldata.calcite.models.ArtifactManager;
import com.rilldata.calcite.models.ArtifactType;
import com.rilldata.calcite.models.InMemoryArtifactManager;
import com.rilldata.calcite.visitors.MetricsViewExpander;
import com.rilldata.protobuf.SqlNodeProtoBuilder;
import org.apache.calcite.config.CalciteConnectionConfigImpl;
import org.apache.calcite.config.CalciteConnectionProperty;
import org.apache.calcite.plan.Context;
import org.apache.calcite.prepare.PlannerImpl;
import org.apache.calcite.schema.SchemaPlus;
import org.apache.calcite.sql.SqlDialect;
import org.apache.calcite.sql.SqlLiteral;
import org.apache.calcite.sql.SqlNode;
import org.apache.calcite.sql.SqlNodeList;
import org.apache.calcite.sql.SqlSelect;
import org.apache.calcite.sql.ddl.SqlCreateTable;
import org.apache.calcite.sql.ddl.SqlCreateView;
import org.apache.calcite.sql.parser.SqlParseException;
import org.apache.calcite.sql.parser.SqlParser;
import org.apache.calcite.sql.parser.SqlParserPos;
import org.apache.calcite.sql.validate.SqlConformanceEnum;
import org.apache.calcite.sql.validate.SqlValidator;
import org.apache.calcite.sql2rel.SqlToRelConverter;
import org.apache.calcite.tools.FrameworkConfig;
import org.apache.calcite.tools.Frameworks;
import org.apache.calcite.tools.Planner;
import org.apache.calcite.tools.ValidationException;
import org.apache.calcite.util.SourceStringReader;
import org.checkerframework.checker.nullness.qual.Nullable;

import java.lang.reflect.Field;
import java.util.ArrayList;
import java.util.Collections;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.Objects;
import java.util.Properties;
import java.util.function.Function;
import java.util.function.Supplier;
import java.util.stream.Collectors;

/**
 * Run `mvn package` to generate the custom SQL parser {@link com.rilldata.calcite.generated.RillSqlParserImpl}
 * under `target/generated-sources/javacc` folder
 * */
public class CalciteToolbox
{
  static final SqlParser.Config PARSER_CONFIG = SqlParser
      .config()
      .withCaseSensitive(false)
      .withParserFactory(RillSqlParserImpl::new);

  private final FrameworkConfig frameworkConfig;
  private final SqlDialect sqlDialect;
  private final ArtifactManager artifactManager;

  public CalciteToolbox(Supplier<SchemaPlus> rootSchemaSupplier, SqlDialect sqlDialect, @Nullable ArtifactManager artifactManager)
  {
    this.sqlDialect = sqlDialect;
    this.artifactManager = Objects.requireNonNullElseGet(artifactManager, InMemoryArtifactManager::new);

    /* Creating CalciteConnectionConfigImpl just like it is done in calcite code but adding LENIENT conformance instead
     of DEFAULT one which does not allow numbers in group by clause like GROUP BY 1,2 ...

     It is kind of odd that validator conformance level is reset to CalciteConnectionConfig level upon creation of validator
     in PlannerImpl#createSqlValidator method
     */
    Properties properties = new Properties();
    properties.setProperty(CalciteConnectionProperty.CASE_SENSITIVE.camelName(), String.valueOf(false));
    properties.setProperty(CalciteConnectionProperty.CONFORMANCE.camelName(), SqlConformanceEnum.LENIENT.name());
    CalciteConnectionConfigImpl config = new CalciteConnectionConfigImpl(properties);

    frameworkConfig = Frameworks
        .newConfigBuilder()
        .defaultSchema(rootSchemaSupplier.get())
        .parserConfig(PARSER_CONFIG)
        .sqlValidatorConfig(SqlValidator.Config.DEFAULT.withConformance(SqlConformanceEnum.LENIENT))
        .context(new Context()
        {
          @Override
          public <C> @Nullable C unwrap(Class<C> aClass)
          {
            if (aClass == CalciteConnectionConfigImpl.class) {
              return aClass.cast(config);
            }
            return null;
          }
        })
        .sqlToRelConverterConfig(SqlToRelConverter.config())
        .build();
  }

  // public as used in unit testing
  public Planner getPlanner()
  {
    return Frameworks.getPlanner(frameworkConfig);
  }

  private static SqlParser getParser(String sql) {
    SqlParser.Config parserConfig = SqlParser
        .config()
        .withCaseSensitive(false)
        .withConformance(SqlConformanceEnum.BABEL)
        .withParserFactory(RillSqlParserImpl::new);

    return SqlParser.create(new SourceStringReader(sql), parserConfig);
  }

  public static SqlNode parseStmt(String sql) throws SqlParseException {
    return getParser(sql).parseStmt();
  }

  public static SqlNode parseStmts(String sql) throws SqlParseException {
    return getParser(sql).parseStmtList();
  }

  public String getRunnableQuery(String sql) throws SqlParseException, ValidationException
  {
    Planner planner = getPlanner();
    SqlNode sqlNode = planner.parse(sql);
    // expand query if needed
    sqlNode = sqlNode.accept(new MetricsViewExpander(artifactManager, this));
    sql = sqlNode.toSqlString(sqlDialect).getSql();
    // expansion done, now validate query
    SqlNode validated = planner.validate(sqlNode);
    planner.close();
    return sql;
  }

  public byte[] getAST(String sql, boolean addTypeInfo) throws SqlParseException, ValidationException
  {
    Planner planner = getPlanner();
    SqlNode sqlNode = planner.parse(sql);
    // expand query if needed
    sqlNode = sqlNode.accept(new MetricsViewExpander(artifactManager, this));
    SqlValidator sqlValidator = null;
    if (addTypeInfo) {
      SqlNode toValidate = sqlNode.clone(sqlNode.getParserPosition());
      planner.validate(toValidate);
      try {
        sqlValidator = getValidator((PlannerImpl) planner);
      } catch (NoSuchFieldException | IllegalAccessException e) {
        throw new RuntimeException(e);
      }
    }
    SqlNodeProtoBuilder sqlNodeProtoBuilder = new SqlNodeProtoBuilder(sqlNode, sqlValidator);
    byte[] bytes = sqlNodeProtoBuilder.getProto();
    planner.close();
    return bytes;
  }

  public SqlValidator getValidator(PlannerImpl planner) throws NoSuchFieldException, IllegalAccessException
  {
    Field validatorField = PlannerImpl.class.getDeclaredField("validator");
    validatorField.setAccessible(true);
    return (SqlValidator) validatorField.get(planner);
  }

  static class Statement
  {
    public List<Statement> dependencies;
    SqlNode node;
    String ddl;
    String name;
    boolean changed;

    public String getName()
    {
      return name;
    }

    public String getType()
    {
      if (node instanceof SqlCreateTable) {
        return "TABLE";
      } else if (node instanceof SqlCreateView) {
        return "VIEW";
      } else if (node instanceof SqlCreateMetric) {
        return "METRICS VIEW";
      }
      return null;
    }
  }

  /**
   * Basically the algorithm should accomplish:
   * 1. Compare statement set 1 (representing the existing state) and statement set 2 (representing the new state).
   * 2. Create DROP statements for entities that either no longer exist in the new state or have their statement changed.
   * 3. Add CREATE statements for entities that do not exist in the existing state.
   *
   * Simplistic algorithm steps:
   *   for each statement in the new state:
   *     if existingState.contains(statement):
   *       existingStatement = existingState.remove(statement)
   *       if statement != existingStatement:
   *         add DROP
   *         add CREATE
   *     if !existingState.contains(statement):
   *       add CREATE
   *  for each statement in the existing state:
   *    add DROP
   *
   * Right now the algorithm doesn't track name changes.
   * The algorithm constructs a graph of dependencies between statements but doesn't use it for now. The dependency graph
   * is required because if a dependency changes its dependant can have a statement unchanged but the entity should be
   * recreated still.
   */
  public static List<MigrationStep> inferMigrations(String newSql, String schema, SqlDialect sqlDialect)
      throws SqlParseException, JsonProcessingException
  {
    SqlNode node = parseStmts(newSql);

    Map<String, Statement> existing = existingStatementsMap(schema, sqlDialect);
    createGraph(existing); // todo use graph to track dependency changes
    Map<String, Statement> ast = newToStatementsMap(node, sqlDialect);

    List<MigrationStep> createSteps = new ArrayList<>();
    List<MigrationStep> dropSteps = new ArrayList<>();
    for (Statement create : ast.values()) {
      Statement migrateFrom = existing.get(create.name);
      if (migrateFrom != null) {
        create.changed = !create.ddl.equals(migrateFrom.ddl);
        if (create.changed) {
          dropSteps.add(MigrationStep.dropEntity(migrateFrom.name, migrateFrom.getType()));
          if (create.getType().equals("METRICS VIEW")) {
            createSteps.add(MigrationStep.insertCatalog(create.ddl));
          } else {
            createSteps.add(MigrationStep.fromDdl(create.ddl));
          }
        }
        existing.remove(create.name); // remove from existing so that we can drop the rest
      } else {
        if (create.getType().equals("METRICS VIEW")) {
          createSteps.add(MigrationStep.insertCatalog(create.ddl));
        } else {
          createSteps.add(MigrationStep.fromDdl(create.ddl));
        }
      }
    }
    for (Statement drop : existing.values()) {
      if (drop.ddl != null) { // skip entities that were not created by the system
        dropSteps.add(MigrationStep.dropEntity(drop.name, drop.getType()));
      }
    }
    int initialCapacity = dropSteps.size() + createSteps.size();
    if (initialCapacity != 0) {
      List<MigrationStep> steps = new ArrayList<>(initialCapacity);
      String dropSql = dropSteps.stream().map(s -> s.ddl).collect(Collectors.joining(";"));
      SqlNodeList sqlNode = (SqlNodeList) parseStmts(dropSql);
      steps.addAll(sqlNode.stream().map(n -> n.toSqlString(sqlDialect).getSql()).map(s -> MigrationStep.fromDdl(s)).collect(Collectors.toList()));
      steps.addAll(createSteps);
      return steps;
    }
    return Collections.emptyList();
  }

  public String saveModel(String sql) throws SqlParseException, ValidationException
  {
    SqlCreateMetric sqlCreateMetric = parseModelingQuery(sql);
    String metricViewString = validateModelingQuery(sqlCreateMetric);
    artifactManager.saveArtifact(
        new Artifact(ArtifactType.METRIC_VIEW, sqlCreateMetric.name.getSimple(), metricViewString));
    return metricViewString;
  }

  public SqlCreateMetric parseModelingQuery(String sql) throws SqlParseException
  {
    Planner planner = getPlanner();
    SqlNode sqlNode = planner.parse(sql);
    planner.close();
    return (SqlCreateMetric) sqlNode;
  }

  /**
   * Validates create metrics view query by parsing and validating group by queries
   * created from the dimensions and measures specified in the modeling query
   */
  public String validateModelingQuery(SqlCreateMetric sqlCreateMetric) throws SqlParseException, ValidationException
  {
    SqlNodeList dimensions = sqlCreateMetric.dimensions;
    SqlNodeList groupByList = new SqlNodeList(SqlParserPos.ZERO);
    for (int i = 1; i <= dimensions.size(); i++) {
      groupByList.add(SqlLiteral.createExactNumeric(i + "", SqlParserPos.ZERO));
    }
    for (SqlNode measure : sqlCreateMetric.measures.getList()) {
      Planner planner = getPlanner();
      SqlNodeList selectList = new SqlNodeList(dimensions, SqlParserPos.ZERO);
      selectList.add(measure);
      SqlSelect groupBy = new SqlSelect(
          SqlParserPos.ZERO,
          SqlNodeList.EMPTY,
          selectList,
          sqlCreateMetric.from,
          null,
          groupByList,
          null,
          null,
          null,
          null,
          null,
          null
      );
      String sqlString = groupBy.toSqlString(sqlDialect).toString();
      SqlNode parsed = planner.parse(sqlString);
      SqlNode validated = planner.validate(parsed);
      planner.close();
    }
    // dimensions, measures and table are validated, return sql string to be stored in db
    return sqlCreateMetric.toSqlString(sqlDialect).toString();
  }

  private static void createGraph(Map<String, Statement> existing)
  {
    for (Statement create : existing.values()) {
      List<String> dependencies = create.node.accept(new DependencyFinder());
      create.dependencies = dependencies.stream().map(existing::get).collect(Collectors.toList());
    }
  }

  private static Map<String, Statement> newToStatementsMap(SqlNode node, SqlDialect sqlDialect)
  {
    SqlNodeList list = (SqlNodeList) node;
    return list.stream().map(n -> {
      Statement statement = new Statement();
      if (n instanceof SqlCreateTable t) {
        statement.name = t.name.toString();
        statement.ddl = n.toSqlString(sqlDialect).toString();
        statement.node = n;
      } else if (n instanceof SqlCreateView v) {
        statement.name = v.name.toString();
        statement.ddl = n.toSqlString(sqlDialect).toString();
        statement.node = n;
      } else if (n instanceof SqlCreateMetric v) {
        statement.name = v.name.toString();
        statement.ddl = n.toSqlString(sqlDialect).toString();
        statement.node = n;
      }
      return statement;
    }).collect(Collectors.toMap(Statement::getName, Function.identity(), (x, y) -> y, LinkedHashMap::new));
  }

  private static Map<String, Statement> existingStatementsMap(String str, SqlDialect dialect) throws JsonProcessingException
  {
    JsonSchema schema = getObjectMapper().readValue(str, JsonSchema.class);
    return schema.entities.stream().map(entity -> {
      Statement statement = new Statement();
      statement.name = entity.name;
      statement.ddl = entity.ddl;
      try {
        statement.node = parseStmt(statement.ddl);
        statement.ddl = statement.node.toSqlString(dialect).toString();
        if (statement.node instanceof SqlCreateView view) {
          statement.name = view.name.toString();
        } else if (statement.node instanceof SqlCreateTable tableNode) {
          statement.name = tableNode.name.toString();
        }
      }
      catch (SqlParseException e) {
        throw new RuntimeException(e);
      }
      return statement;
    }).collect(Collectors.toMap(Statement::getName, Function.identity(), (a, b) -> b, LinkedHashMap::new));
  }

  private static ObjectMapper getObjectMapper()
  {
    return new ObjectMapper().configure(DeserializationFeature.FAIL_ON_UNKNOWN_PROPERTIES, false);
  }
}
