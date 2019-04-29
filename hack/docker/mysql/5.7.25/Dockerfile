FROM mysql:5.7.25

COPY on-start.sh /
COPY peer-finder /usr/local/bin/

RUN chmod +x /on-start.sh

# For standalone mysql
# default entrypoint of parent mysql:5.7.25
# ENTRYPOINT ["docker-entrypoint.sh"]

# For mysql group replication
# ENTRYPOINT ["peer-finder"]
