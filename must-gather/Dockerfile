FROM registry.redhat.io/openshift4/ose-must-gather

# Save original gather script
COPY /usr/bin/gather /usr/bin/gather_original

# Copy all scripts to /usr/bin
COPY gather_appsodyoperator/* /usr/bin/

ENTRYPOINT /usr/bin/gather
