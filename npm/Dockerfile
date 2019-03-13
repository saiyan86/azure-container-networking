# Use a minimal image as a parent image
FROM ubuntu:18.10
ARG NPM_BUILD_DIR
ARG SCRIPTS_DIR

# Install dependencies.
COPY $SCRIPTS_DIR/install-npm-dependencies.sh .
RUN chmod +x ./install-npm-dependencies.sh
RUN ./install-npm-dependencies.sh > /dev/null

# Install plugin.
COPY $NPM_BUILD_DIR/azure-npm /usr/bin
WORKDIR /usr/bin

# Run the npm command by default when the container starts.
ENTRYPOINT ["/usr/bin/azure-npm"]