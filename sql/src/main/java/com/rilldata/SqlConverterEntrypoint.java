package com.rilldata;

import com.rilldata.calcite.dialects.Dialects;
import com.rilldata.protobuf.generated.Requests;
import com.rilldata.protobuf.generated.SqlNodeProto;
import org.graalvm.nativeimage.IsolateThread;
import org.graalvm.nativeimage.c.function.CEntryPoint;
import org.graalvm.nativeimage.c.function.CFunctionPointer;
import org.graalvm.nativeimage.c.function.InvokeCFunctionPointer;
import org.graalvm.nativeimage.c.type.CCharPointer;
import org.graalvm.nativeimage.c.type.CTypeConversion;
import org.graalvm.word.WordFactory;

import java.io.PrintWriter;
import java.io.StringWriter;
import java.util.Base64;

/**
 * This class contains an entry point (a function callable from a native executable, ie C/Go executable).
 */
public class SqlConverterEntrypoint
{
//  private static SqlConverter sqlConverter;

  interface AllocatorFn extends CFunctionPointer
  {
    @InvokeCFunctionPointer
    CCharPointer call(long size);
  }

  public static Requests.Response transpile(Requests.Request r) {
    Requests.TranspileRequest transpileRequest = r.getTranspileRequest();
    String sql = transpileRequest.getSql();
    Requests.Dialect dialect = transpileRequest.getDialect();
    try {
      SqlConverter sqlConverter = new SqlConverter(transpileRequest.getCatalog());
      String transpiledSql = sqlConverter.convert(sql, Dialects.valueOf(dialect.name()).getSqlDialect());
      return Requests.Response
          .newBuilder()
          .setTranspileResponse(Requests.TranspileResponse.newBuilder().setSql(transpiledSql).build())
          .build();
    } catch (Exception e) {
      return Requests.Response
          .newBuilder()
          .setError(
            Requests.Error.newBuilder().setMessage(e.getMessage()).setStackTrace(stackTraceToString(e)).build())
          .build();
    }
  }

  public static Requests.Response parse(Requests.Request r) {
    Requests.ParseRequest parseRequest = r.getParseRequest();
    String sql = parseRequest.getSql();
    Requests.Response response;
    try {
      SqlConverter sqlConverter = new SqlConverter(parseRequest.getCatalog());
      SqlNodeProto sqlNodeProto = sqlConverter.getAST(sql, parseRequest.getAddTypeInfo());
      response = Requests.Response
          .newBuilder()
          .setParseResponse(Requests.ParseResponse.newBuilder().setAst(sqlNodeProto).build())
          .build();
    } catch (Exception e) {
      response = Requests.Response
          .newBuilder()
          .setError(
              Requests.Error.newBuilder().setMessage(e.getMessage()).setStackTrace(stackTraceToString(e)).build())
          .build();
    }
    return response;
  }

  @CEntryPoint(name = "request")
  public static CCharPointer processRequest(IsolateThread thread, AllocatorFn allocatorFn, CCharPointer request) {
    String b64String = CTypeConversion.toJavaString(request);
    byte[] decoded = Base64.getDecoder().decode(b64String);

    try {
      Requests.Request r = Requests.Request.parseFrom(decoded);
      if (r.hasParseRequest()) {
        byte[] b64response = Base64.getEncoder().encode(parse(r).toByteArray());
        return convertToCCharPointer(allocatorFn, b64response);
      } else if (r.hasTranspileRequest()) {
          byte[] response = transpile(r).toByteArray();
          byte[] b64response = Base64.getEncoder().encode(response);
          return convertToCCharPointer(allocatorFn, b64response);
      }
      Requests.Response build = Requests.Response
          .newBuilder()
          .setError(Requests.Error.newBuilder().setMessage("Empty request").build())
          .build();
      
      byte[] b64response = Base64.getEncoder().encode(build.toByteArray());
      return convertToCCharPointer(allocatorFn, b64response);
    } catch (Exception e) {
      Requests.Response build = Requests.Response
          .newBuilder()
          .setError(
              Requests.Error.newBuilder().setMessage(e.getMessage()).setStackTrace(stackTraceToString(e)).build())
          .build();
      byte[] b64response = Base64.getEncoder().encode(build.toByteArray());
      return convertToCCharPointer(allocatorFn, b64response);
    }
  }

  @CEntryPoint(name = "convert_sql")
  public static CCharPointer convertSql(IsolateThread thread, AllocatorFn allocatorFn, CCharPointer sql,
      CCharPointer catalog, CCharPointer dialect
  )
  {
    try {
      String dialectString = CTypeConversion.toJavaString(dialect);
      Dialects dialectEnum = Dialects.valueOf(dialectString.toUpperCase());
      String javaCatalogString = CTypeConversion.toJavaString(catalog);
      SqlConverter sqlConverter = new SqlConverter(javaCatalogString);
      String javaSqlString = CTypeConversion.toJavaString(sql);
      String runnableQuery = sqlConverter.convert(javaSqlString, dialectEnum.getSqlDialect());
      if (runnableQuery == null) {
        return WordFactory.nullPointer();
      }
      return convertToCCharPointer(allocatorFn, runnableQuery);
    } catch (Exception e) {
      e.printStackTrace(); // todo level-logging for native libraries?
      return convertToCCharPointer(allocatorFn, String.format("{'error': '%s'}", e.getMessage()));
    }
  }

  @CEntryPoint(name = "get_ast")
  public static CCharPointer getAST(IsolateThread thread, AllocatorFn allocatorFn, CCharPointer sql,
      CCharPointer catalog
  )
  {
    try {
      String javaCatalogString = CTypeConversion.toJavaString(catalog);
      SqlConverter sqlConverter = new SqlConverter(javaCatalogString);
      String sqlString = CTypeConversion.toJavaString(sql);
      SqlNodeProto ast = sqlConverter.getAST(sqlString);
      return convertToCCharPointer(allocatorFn, ast.toByteArray());
    } catch (Exception e) {
      e.printStackTrace();
      return WordFactory.nullPointer();
    }
  }

  private static CCharPointer convertToCCharPointer(AllocatorFn allocatorFn, String javaString)
  {
    return convertToCCharPointer(allocatorFn, javaString.getBytes());
  }

  private static CCharPointer convertToCCharPointer(AllocatorFn allocatorFn, byte[] b)
  {
    CCharPointer a = allocatorFn.call(b.length + 1);
    for (int i = 0; i < b.length; i++) {
      a.write(i, b[i]);
    }
    a.write(b.length, (byte) 0);
    return a;
  }

  private static String stackTraceToString(Exception e)
  {
    StringWriter sw = new StringWriter();
    PrintWriter pw = new PrintWriter(sw);
    e.printStackTrace(pw);
    return sw.toString();
  }
}
