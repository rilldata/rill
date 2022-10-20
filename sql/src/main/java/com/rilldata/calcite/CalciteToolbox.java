package com.rilldata.calcite;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.rilldata.StaticSchemaProvider;
import com.rilldata.calcite.dialects.Dialects;
import com.rilldata.calcite.generated.RillSqlParserImpl;
import com.rilldata.calcite.models.SqlCreateMetricsView;
import com.rilldata.calcite.models.SqlCreateSource;
import com.rilldata.calcite.models.ArtifactStore;
import com.rilldata.calcite.operators.RillOperatorTable;
import com.rilldata.calcite.validators.CreateMetricsViewValidator;
import com.rilldata.calcite.validators.CreateSourceValidator;
import com.rilldata.calcite.visitors.MetricsViewExpander;
import com.rilldata.protobuf.SqlNodeProtoBuilder;
import org.apache.calcite.config.CalciteConnectionConfigImpl;
import org.apache.calcite.config.CalciteConnectionProperty;
import org.apache.calcite.plan.Context;
import org.apache.calcite.prepare.PlannerImpl;
import org.apache.calcite.schema.SchemaPlus;
import org.apache.calcite.sql.SqlDialect;
import org.apache.calcite.sql.SqlNode;
import org.apache.calcite.sql.parser.SqlParseException;
import org.apache.calcite.sql.parser.SqlParser;
import org.apache.calcite.sql.validate.SqlConformanceEnum;
import org.apache.calcite.sql.validate.SqlValidator;
import org.apache.calcite.tools.FrameworkConfig;
import org.apache.calcite.tools.Frameworks;
import org.apache.calcite.tools.Planner;
import org.apache.calcite.tools.ValidationException;
import org.checkerframework.checker.nullness.qual.Nullable;

import java.io.IOException;
import java.lang.reflect.Field;
import java.util.Objects;
import java.util.Properties;
import java.util.function.Supplier;

/**
 * Run `mvn package` to generate the custom SQL parser {@link com.rilldata.calcite.generated.RillSqlParserImpl}
 * under `target/generated-sources/javacc` folder
 */
public class CalciteToolbox
{
  static final SqlParser.Config PARSER_CONFIG = SqlParser.config().withCaseSensitive(false)
      .withParserFactory(RillSqlParserImpl::new);

  private final FrameworkConfig frameworkConfig;
  private final ArtifactStore artifactStore;

  public static CalciteToolbox buildToolbox(String catalog) throws IOException
  {
    JsonCatalog jsonCatalog = new ObjectMapper().readValue(catalog, JsonCatalog.class);
    return new CalciteToolbox(new StaticSchemaProvider(jsonCatalog.schemas),
        new ArtifactStore(jsonCatalog.artifacts)
    );
  }

  public CalciteToolbox(Supplier<SchemaPlus> rootSchemaSupplier, @Nullable ArtifactStore artifactStore)
  {
    this.artifactStore = Objects.requireNonNullElseGet(artifactStore, ArtifactStore::new);
    /*
      Creating CalciteConnectionConfigImpl just like it is done in calcite code but adding LENIENT
      conformance instead of DEFAULT one which does not allow numbers in group by clause like GROUP BY 1,2 ...
      It is kind of odd that validator conformance level is reset to CalciteConnectionConfig level upon creation
      of validator in PlannerImpl#createSqlValidator method
     */
    Properties properties = new Properties();
    properties.setProperty(CalciteConnectionProperty.CASE_SENSITIVE.camelName(), String.valueOf(false));
    properties.setProperty(CalciteConnectionProperty.CONFORMANCE.camelName(), SqlConformanceEnum.LENIENT.name());
    CalciteConnectionConfigImpl config = new CalciteConnectionConfigImpl(properties);

    frameworkConfig = Frameworks.newConfigBuilder()
        .defaultSchema(rootSchemaSupplier.get())
        .parserConfig(PARSER_CONFIG)
        .sqlValidatorConfig(
            SqlValidator.Config.DEFAULT.withTypeCoercionEnabled(false).withConformance(SqlConformanceEnum.LENIENT))
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
        .operatorTable(new RillOperatorTable())
        .build();
  }

  // public as used in unit testing
  public Planner getPlanner()
  {
    return Frameworks.getPlanner(frameworkConfig);
  }

  public String getRunnableQuery(String sql, SqlDialect sqlDialect) throws SqlParseException, ValidationException
  {
    Planner planner = getPlanner();
    SqlNode sqlNode = planner.parse(sql);
    // expand query if needed
    sqlNode = sqlNode.accept(new MetricsViewExpander(artifactStore, this));
    // expansion done, now validate query
    SqlNode validated = planner.validate(sqlNode);
    planner.close();
    sql = validated.toSqlString(sqlDialect).getSql();
    return sql;
  }

  public byte[] getAST(String sql, boolean addTypeInfo) throws SqlParseException, ValidationException
  {
    Planner planner = getPlanner();
    SqlNode sqlNode = planner.parse(sql);
    if (sqlNode instanceof SqlCreateMetricsView sqlCreateMetricsView) {
      CreateMetricsViewValidator.validateModelingQuery(sqlCreateMetricsView, Dialects.DUCKDB.getSqlDialect(),
          getPlanner()
      );
    } else if (sqlNode instanceof SqlCreateSource sqlCreateSource) {
      CreateSourceValidator.validateConnector(sqlCreateSource);
    }
    return getAST(sqlNode, planner, addTypeInfo);
  }

  public byte[] getAST(SqlNode sqlNode)
  {
    Planner planner = getPlanner();
    return getAST(sqlNode, planner, false);
  }

  /**
   * If addTypeInfo is true then the same planner passed here should have been used to parse the sql
   * otherwise the sql validation will fail which is required to get type info
   */
  public byte[] getAST(SqlNode sqlNode, Planner planner, boolean addTypeInfo)
  {
    SqlValidator sqlValidator = null;
    if (addTypeInfo) {
      SqlNode toValidate = sqlNode.clone(sqlNode.getParserPosition());
      try {
        planner.validate(toValidate);
        sqlValidator = getValidator((PlannerImpl) planner);
      } catch (NoSuchFieldException | IllegalAccessException | ValidationException e) {
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

  public SqlNode parseValidatedSql(String sql) throws SqlParseException
  {
    Planner planner = getPlanner();
    SqlNode sqlNode = planner.parse(sql);
    planner.close();
    return sqlNode;
  }
}
