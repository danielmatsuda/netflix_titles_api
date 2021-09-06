# go_rest_api

## About my project

List of features:
- REST API with CRUD functionality and filtering
- HTTPS (via Caddy/Let's Encrypt)
- error handling (input validation, server error handling)
- connected to PostgreSQL database; database import via SQL migrations
- utilizes concurrency safely via httprouter (net/http) goroutines
- middleware - rate limiting of global requests, panic recovery, API usage metrics
- handles context timeouts and request queueing (via Go's sql.DB connection pool)

## Things I learned

By doing this project, I learned a lot about API development and Go. Because I learned a lot, there are many comments in my code to remind me exactly what each chunk of code is doing.

Things I learned about:
- easier routing with httprouter
- querying PostgreSQL in Go (i.e., pq and sql packages)
- connecting my API to a PostgreSQL database, and doing SQL migrations (migrate package)
- using custom functions to alter a package's default behavior (i.e., for customizing errors and JSON reading/writing)
- refresher on pointers and dereferencers (& and *), and how to use them in functions
- writing an input validation suite
- writing idiomatic middleware functions in Go (i.e., using closures)
- implementing HTTPS/reverse proxy using Caddy

## Acknowledgments

Much of this repo's code is heavily based off of Alex Edwards' [Let's Go Further](https://lets-go-further.alexedwards.net/), which I referenced often as I created my API. 

I used the book to learn about a REST API's key components, and how to implement them using idiomatic Go code and useful third-party packages. Unlike the book, I created an API for my own data source, and hosted the API on an AWS EC2 instance. I also referenced materials from my school's Cloud App Development course, package docs, and various online tutorials.

The API's data was taken from [this Kaggle dataset](https://www.kaggle.com/shivamb/netflix-shows) in August 2021.

## Using the API

My REST API has the following CRUD functionality:
1. GET a single title by id. (v1/titles/:id)
2. GET all entries or a filtered set of entries, using query string parameters to filter. (v1/titles)
3. Create (POST) a single title by providing all fields except for ID. (v1/titles)
4. Update (PUT) one or more fields of a single title (except for the ID field). The update is called on all fields of the entry (besides ID), so the client must provide valid values for all fields. (v1/titles/:id)
5. DELETE a single title entry. (v1/titles/:id)