FROM migrate/migrate

COPY ./entrypoint.sh /app/do_migrate.sh
RUN chmod +x /app/do_migrate.sh

COPY ./data/postgres /migrations

ENTRYPOINT []
CMD /app/do_migrate.sh
