FROM openjdk:11

ARG DRUID_VERSION=0.23.0
ENV DRUID_VERSION ${DRUID_VERSION}
ENV LOG_LEVEL info

RUN apt-get update \
      && apt-get install -y curl gettext-base

# Install druid
RUN curl https://downloads.apache.org/druid/${DRUID_VERSION}/apache-druid-${DRUID_VERSION}-bin.tar.gz > /opt/druid-${DRUID_VERSION}-bin.tar.gz \
      && tar -xvf /opt/druid-${DRUID_VERSION}-bin.tar.gz -C /opt/ \
      && rm -f /opt/druid-${DRUID_VERSION}-bin.tar.gz \
      && mv /opt/apache-druid-${DRUID_VERSION} /opt/druid \
      && mkdir -p /var/log/druid \
      && mkdir -p /opt/druid/data 

WORKDIR /opt/druid

# Remove unneeded files
RUN rm -rf $WORKDIR/extensions
RUN rm -rf $WORKDIR/hadoop-dependencies
RUN rm -rf $WORKDIR/quickstart

# coordinator/overlord
EXPOSE 8081
# broker
EXPOSE 8082
# historical
EXPOSE 8083
# router
EXPOSE 8888
# middle manager
EXPOSE 8091

CMD ./bin/start-micro-quickstart
