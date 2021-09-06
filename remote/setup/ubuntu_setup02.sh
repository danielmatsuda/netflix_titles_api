# Sets up the netflix_db and performs SQL migrations on the local machine. 

# create a db user in Postgres
DB_ROLE=netflix_db
DB_NAME=netflix_db
read -p "Choose a password for db user netflix_db: " DB_PASSWORD

# Log in so that passwords aren't required for the following commands
#sudo -i -u postgres psql postgres
# --- now, type \password postgres here to manually set a new password for root ---- #######

sudo -i -u postgres psql -c "CREATE DATABASE ${DB_NAME}"
sudo -i -u postgres psql -d ${DB_NAME} -c "CREATE EXTENSION IF NOT EXISTS citext"
sudo -i -u postgres psql -d ${DB_NAME} -c "CREATE ROLE ${DB_ROLE} WITH LOGIN PASSWORD '${DB_PASSWORD}'"
# Add a DSN as a system-wide env, to be used by the Go app
NETFLIX_DB_DSN="postgres://${DB_ROLE}:${DB_PASSWORD}@localhost/${DB_NAME}"
sudo bash -c "echo ${NETFLIX_DB_DSN} >> /etc/environment"

# perform sql migrations, including giving permissions needed for COPY
sudo -i -u postgres psql -d ${DB_NAME} -c "GRANT pg_read_server_files TO ${DB_ROLE};"
sudo migrate -path=./migrations -database=${NETFLIX_DB_DSN} up

# finally, run the API as a background process
nohup sudo ./api -port=4000 -db-dsn=$NETFLIX_DB_DSN -env=production &