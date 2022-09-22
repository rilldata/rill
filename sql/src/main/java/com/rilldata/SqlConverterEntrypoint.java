package com.rilldata;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.rilldata.calcite.CalciteToolbox;
import com.rilldata.calcite.MigrationStep;
import org.apache.calcite.sql.dialect.PostgresqlSqlDialect;
import org.graalvm.nativeimage.IsolateThread;
import org.graalvm.nativeimage.c.function.CEntryPoint;
import org.graalvm.nativeimage.c.function.CFunctionPointer;
import org.graalvm.nativeimage.c.function.InvokeCFunctionPointer;
import org.graalvm.nativeimage.c.type.CCharPointer;
import org.graalvm.nativeimage.c.type.CTypeConversion;
import org.graalvm.word.WordFactory;

import java.util.List;

/**
 * This class contains an entry point (a function callable from a native executable, ie C/Go executable).
 */
public class SqlConverterEntrypoint
{
  private static SqlConverter sqlConverter;

  interface AllocatorFn extends CFunctionPointer
  {
    @InvokeCFunctionPointer
    CCharPointer call(long size);
  }

  @CEntryPoint(name="convert_sql")
  public static CCharPointer convertSql(IsolateThread thread, AllocatorFn allocatorFn, CCharPointer sql, CCharPointer schema)
  {
    try {
      String javaSchemaString = CTypeConversion.toJavaString(schema);
      SqlConverter sqlConverter = new SqlConverter(javaSchemaString);
      String javaSqlString = CTypeConversion.toJavaString(sql);
      String runnableQuery = sqlConverter.convert(javaSqlString);
      if (runnableQuery == null) {
        return WordFactory.nullPointer();
      }
      return convertToCCharPointer(allocatorFn, runnableQuery);
    } catch (Exception e) {
      e.printStackTrace();
      return WordFactory.nullPointer();
    }
  }

  @CEntryPoint(name="apply")
  public static CCharPointer inferMigrationsSteps(IsolateThread thread, AllocatorFn allocatorFn, CCharPointer json, CCharPointer catalog)
  {
    try {
      String javaSchemaString = CTypeConversion.toJavaString(catalog);
      String javaSqlString = CTypeConversion.toJavaString(json);
      List<MigrationStep> migrationSteps = CalciteToolbox.inferMigrations(
          javaSqlString,
          javaSchemaString,
          PostgresqlSqlDialect.DEFAULT
      );
      String stepsJson = new ObjectMapper().writeValueAsString(migrationSteps);
      return convertToCCharPointer(allocatorFn, stepsJson);
    } catch (Exception e) {
      e.printStackTrace();
      return WordFactory.nullPointer();
    }
  }

  private static CCharPointer convertToCCharPointer(AllocatorFn allocatorFn, String javaString)
  {
    byte[] b = javaString.getBytes();
    CCharPointer a =  allocatorFn.call(b.length + 1);
    for (int i = 0; i < b.length; i++) {
      a.write(i, b[i]);
    }
    a.write(b.length, (byte) 0);
    return a;
  }
}
