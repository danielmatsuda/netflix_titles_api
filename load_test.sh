touch load_test.txt
# Send 30 GET requests per second for 60 seconds for random titles in the database
# (may include queries for non-existent IDs between ~7700-9999)
pewpew benchmark -d=60 --rps=30 -X=GET -H="Content-Type:application/json" \
-r "https://${NETFLIX_API_HOSTNAME}/v1/titles/[0-9]{1,4}" > load_test.txt