FROM quay.io/openshift/origin-must-gather:4.2

# Save original gather script
RUN mv /usr/bin/gather /usr/bin/gather_original

# Copy all scripts to /usr/bin
COPY gather_appsodyoperator/* /usr/bin/

ENTRYPOINT /usr/bin/gather
