FROM alpine:latest
# Install SQLite
RUN apk --no-cache add sqlite
# Create a directory to store the database
WORKDIR /db
# Copy your SQLite database file into the container
COPY ../sql/schema.sql /db/
# Expose the port if needed
# EXPOSE 1433
# Command to run when the container starts
CMD ["sqlite3", "/data/schema.sql"]
