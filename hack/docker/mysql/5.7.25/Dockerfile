FROM mysql:5.7.25

COPY on-start.sh /
COPY peer-finder /usr/local/bin/

RUN chmod +x /on-start.sh

ENTRYPOINT ["peer-finder"]
