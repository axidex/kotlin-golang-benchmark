####
# Multi-stage build for Quarkus application
####

# Build stage
FROM gradle:8.5-jdk21 AS builder

WORKDIR /app

# Copy gradle files
COPY gradle gradle
COPY gradlew gradlew.bat gradle.properties settings.gradle.kts build.gradle.kts ./

# Download dependencies
RUN ./gradlew --no-daemon dependencies

# Copy source code
COPY src src

# Build the application
RUN ./gradlew --no-daemon build -x test

# Runtime stage
FROM registry.access.redhat.com/ubi8/openjdk-21:1.20

ENV LANGUAGE='en_US:en'

# Copy the built application from builder
COPY --from=builder --chown=185 /app/build/quarkus-app/lib/ /deployments/lib/
COPY --from=builder --chown=185 /app/build/quarkus-app/*.jar /deployments/
COPY --from=builder --chown=185 /app/build/quarkus-app/app/ /deployments/app/
COPY --from=builder --chown=185 /app/build/quarkus-app/quarkus/ /deployments/quarkus/

EXPOSE 8080
USER 185
ENV JAVA_OPTS_APPEND="-Dquarkus.http.host=0.0.0.0 -Djava.util.logging.manager=org.jboss.logmanager.LogManager"
ENV JAVA_APP_JAR="/deployments/quarkus-run.jar"

ENTRYPOINT [ "/opt/jboss/container/java/run/run-java.sh" ]