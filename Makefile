# compile the binary for both my Windows machine and for Ubuntu Linux
.PHONY: build
build:
	go build -o=./bin/api ./cmd/api
	GOOS=linux GOARCH=amd64 go build -ldflags='-s' -o=./bin/linux_amd64/api ./cmd/api

# run the Go program locally
.PHONY: run
run:
	@go run ./cmd/api -db-dsn=${NETFLIX_DB_DSN}

# open psql locally for the db
.PHONY: psql
psql:
	psql ${NETFLIX_DB_DSN}

# run up migrations locally
.PHONY: up
up:
	@echo 'Running up migrations...'
	migrate -path ./migrations -database ${NETFLIX_DB_DSN} up

# API is hosted on the AWS elastic IP I assigned to the EC2 instance
production_host_ip = ${AWS_ELASTIC_IP}

# connect to the EC2 machine
.PHONY: connect
connect:
	ssh -i C:/AWS/EC2/ec2_netflix_api.pem ubuntu@${production_host_ip}

# copy Go app binary (Ubuntu), setup scripts, clean CSV data, sql migrations, and Caddyfile to EC2
.PHONY: setup
setup:
	scp -i C:/AWS/EC2/ec2_netflix_api.pem C:/Users/"Daniel Matsuda"/VSCodeProjects/netflix_db_api/bin/linux_amd64/api ubuntu@${production_host_ip}:/home/ubuntu/
	scp -i C:/AWS/EC2/ec2_netflix_api.pem -r C:/Users/"Daniel Matsuda"/VSCodeProjects/netflix_db_api/migrations ubuntu@${production_host_ip}:/home/ubuntu/
	scp -i C:/AWS/EC2/ec2_netflix_api.pem C:/Users/"Daniel Matsuda"/VSCodeProjects/netflix_db_api/remote/setup/ubuntu_setup01.sh ubuntu@${production_host_ip}:/home/ubuntu/
	scp -i C:/AWS/EC2/ec2_netflix_api.pem C:/Users/"Daniel Matsuda"/VSCodeProjects/netflix_db_api/remote/setup/ubuntu_setup02.sh ubuntu@${production_host_ip}:/home/ubuntu/
	scp -i C:/AWS/EC2/ec2_netflix_api.pem C:/Users/"Daniel Matsuda"/VSCodeProjects/netflix_db_api/migrations/trimmed_netflix_titles.csv ubuntu@${production_host_ip}:/home/ubuntu/
	scp -i C:/AWS/EC2/ec2_netflix_api.pem C:/Users/"Daniel Matsuda"/VSCodeProjects/netflix_db_api/remote/setup/Caddyfile ubuntu@${production_host_ip}:/etc/caddy/

# To run the server setup and sql migrations, I'll have to use make connect, then run the two setup scripts manually (or rewrite the Makefile using Linux bash).
