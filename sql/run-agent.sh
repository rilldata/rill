(
  set -e
  mvn package
  mvn dependency:copy-dependencies
  java -agentlib:native-image-agent=config-merge-dir=graalcfg,experimental-class-define-support -cp target/classes:target/dependency/* com.rilldata.SqlConverterMain sele
)

