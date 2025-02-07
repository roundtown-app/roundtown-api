# roundtown-api
This is the source code for the API that handles virtually all requests.

It will cover the following:
* Events
* Venues
* Plans
* Users
* Recommendations

## Instructions
### Local Testing
Before running the API, [set up a local PostgreSQL instance](https://www.codecademy.com/article/installing-and-using-postgresql-locally) and add [the schema](https://github.com/roundtown-app/roundtown-database/blob/main/roundtown.sql) to it. 

Then, run get_auth_token.py to create test accounts or sign in with them. When you sign in, make sure to copy the token that is printed to the console - this will be needed for requests to the API.

The API itself can be built and run with the following commands:
```
go build cmd/api/main.go
./main
```

I recommend Postman for interacting with the API. When making a request, go to the Authorization tab, select type Bearer token, then paste the token from get_auth_token.py. JSON can be added to the request via the Body tab.

To see the endpoints, go to ./internal/handlers/handler.go
To see the ORM structs, go to ./api

### Deployment
Run the following commands locally:
```
docker build -t api .

docker run -p 8080:8080 api   // for making sure the container image works

docker save api:latest > api[YYYYMMDD-rev].tar   // ex. api20241027-4.tar

scp -i [path to your Lightsail private key] api[YYYYMMDD-rev].tar ec2-user@[server's public IPv4 address]:~
```

Then, log into the server with the following command:
```
ssh -i [path to your Lightsail private key] ec2-user@[server's public IPv4 address]:~
```

Once connected to the server, run the following commands:
```
docker ps -a   // see what containers are running

docker stop [name of running container]   // do this for all running containers

docker system prune -a

docker load < api[YYYYMMDD-rev].tar

docker run -d -p 80:8080 api

docker ps -a   // make sure it's running
```