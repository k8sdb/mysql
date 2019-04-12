FROM mysql:5.7.25

COPY on-change.sh /
COPY peer-finder /usr/local/bin/

RUN chmod +x /on-change.sh

ENTRYPOINT ["peer-finder"]
