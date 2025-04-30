# Docker for erupe

## Building the container
Run the following from the route of the soruce folder. In this example we give it the tag of dev to seperate it from any other container verions. 
```bash
docker build . -t erupe:dev
```
## Running the container in isolation
This is just running the container. You can do volume mounts into the container for the `config.json` to tell it to communicate to a database. You will need to do this also for other folders such as `bin` and `savedata`
```bash
docker run erupe:dev
```

## Docker compose
Docker compose allows you to run multiple containers at once. The docker compose in this folder has 3 things set up.
- postgres
- pg admin (Admin interface to make db changes)
- erupe

We automatically populate the database to the latest version on start. If you you are updating you will need to apply the new schemas manually.

Before we get started you should make sure the database info matches whats in the docker compose file for the environment variables `POSTGRES_PASSWORD`,`POSTGRES_USER` and `POSTGRES_DB`. You can set the host to be the service name `db`.

Here is a example of what you would put in the config.json if you was to leave the defaults. It is strongly recommended to change the password. 
```txt
"Database": {
    "Host": "db",
    "Port": 5432,
    "User": "postgres",
    "Password": "password",
    "Database": "erupe"
  },
```

Place this file within ./docker/config.json

You will need to do the same for your bins place these in ./docker/bin

# Setting up the web hosted materials
Clone the Severs repo into ./docker/Severs

Make sure your hosts are pointing to where this is hosted



## Turning off the server safely
```bash
docker-compose stop
```

## Turning off the server destructive
```bash
docker-compose down
```
Make sure if you want to delete your data you delete the folders that persisted
- ./docker/savedata
- ./docker/db-data
## Turning on the server again 
This boots the db pgadmin and the server in a detached state
```bash
docker-compose up -d
```
if you want all the logs and you want it to be in an attached state 
```bash
docker-compose up
```


# Troubleshooting
Q: My Postgres will not populate. A: You're setup.sh is maybe saved as CRLF it needs to be saved as LF. 