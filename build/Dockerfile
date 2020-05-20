FROM registry.access.redhat.com/ubi8/ubi-minimal:latest

LABEL vendor="Appsody" \
      name="Appsody Application Operator" \
      version="0.6.0" \
      summary="Image for Appsody Application Operator" \
      description="This image contains the controller for Appsody Application Operator. See https://github.com/appsody/appsody-operator#appsody-application-operator"
      
ENV OPERATOR=/usr/local/bin/appsody-operator \
    USER_UID=1001 \
    USER_NAME=appsody-operator

# install operator binary
COPY build/_output/bin/appsody-operator ${OPERATOR}
COPY deploy/stack_defaults.yaml deploy/
COPY deploy/stack_constants.yaml deploy/
COPY build/bin /usr/local/bin
COPY LICENSE /licenses/

RUN  /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}