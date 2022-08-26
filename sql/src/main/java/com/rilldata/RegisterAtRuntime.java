package com.rilldata;

import java.lang.annotation.Annotation;
import java.util.Arrays;
import org.apache.calcite.rel.metadata.*;
import org.apache.calcite.runtime.CalciteContextException;
import org.apache.calcite.runtime.Resources;
import org.apache.calcite.sql.fun.SqlStdOperatorTable;
import org.apache.calcite.sql.validate.SqlValidatorException;
import org.apache.calcite.sql2rel.StandardConvertletTable;
import org.apache.calcite.util.BuiltInMethod;
import org.graalvm.nativeimage.hosted.Feature;
import org.graalvm.nativeimage.hosted.RuntimeReflection;
import org.reflections.Reflections;
import org.reflections.scanners.FieldAnnotationsScanner;
import org.reflections.scanners.SubTypesScanner;

/**
 * This class registers classes that cannot be registered automatically by the `native-image` compiler.
 * see https://github.com/substrait-io/substrait-java/blob/main/isthmus/src/main/java/io/substrait/isthmus/RegisterAtRuntime.java
 * substraite-java compiles Apache Calcite to a native image. It can use some features of Apache Calcite that rill-sql doesn't use that way
 * this file can optimized removing some unused classes but they can be used eventually.
 */
public class RegisterAtRuntime implements Feature {
  public void beforeAnalysis(BeforeAnalysisAccess access) {
    try {
      // calcite items
      Reflections calcite =
          new Reflections(
              "org.apache.calcite", new FieldAnnotationsScanner(), new SubTypesScanner());
      register(BuiltInMetadata.class);
      register(SqlValidatorException.class);
      register(CalciteContextException.class);
      register(SqlStdOperatorTable.class);
      register(StandardConvertletTable.class);
      registerByParent(calcite, Metadata.class);
      registerByParent(calcite, MetadataHandler.class);
      registerByParent(calcite, Resources.Element.class);

      Arrays.asList(
          RelMdPercentageOriginalRows.class,
          RelMdColumnOrigins.class,
          RelMdExpressionLineage.class,
          RelMdTableReferences.class,
          RelMdNodeTypes.class,
          RelMdRowCount.class,
          RelMdMaxRowCount.class,
          RelMdMinRowCount.class,
          RelMdUniqueKeys.class,
          RelMdColumnUniqueness.class,
          RelMdPopulationSize.class,
          RelMdSize.class,
          RelMdParallelism.class,
          RelMdDistribution.class,
          RelMdLowerBoundCost.class,
          RelMdMemory.class,
          RelMdDistinctRowCount.class,
          RelMdSelectivity.class,
          RelMdExplainVisibility.class,
          RelMdPredicates.class,
          RelMdAllPredicates.class,
          RelMdCollation.class)
            .forEach(RegisterAtRuntime::register);

      RuntimeReflection.register(Resources.class);
      RuntimeReflection.register(SqlValidatorException.class);

      Arrays.stream(BuiltInMethod.values())
            .forEach(
                c -> {
                  if (c.field != null) RuntimeReflection.register(c.field);
                  if (c.constructor != null) RuntimeReflection.register(c.constructor);
                  if (c.method != null) RuntimeReflection.register(c.method);
                });
    } catch (Exception e) {
      throw new RuntimeException(e);
    }
  }

  private static void register(Class<?> c) {
    RuntimeReflection.register(c);
    RuntimeReflection.register(c.getDeclaredConstructors());
    RuntimeReflection.register(c.getDeclaredFields());
    RuntimeReflection.register(c.getDeclaredMethods());
    RuntimeReflection.register(c.getConstructors());
    RuntimeReflection.register(c.getFields());
    RuntimeReflection.register(c.getMethods());
  }

  private static void registerByAnnotation(Reflections reflections, Class<? extends Annotation> c) {
    reflections.getTypesAnnotatedWith(c).stream()
               .forEach(
                   inner -> {
                     register(inner);
                     reflections.getSubTypesOf(c).stream().forEach(RegisterAtRuntime::register);
                   });
  }

  private static void registerByParent(Reflections reflections, Class<?> c) {
    register(c);
    reflections.getSubTypesOf(c).stream().forEach(RegisterAtRuntime::register);
  }
}