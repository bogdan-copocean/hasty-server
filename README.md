## Summary
I chose a Microservice architecutre approach for this project instead of a Monolithic architecture for the following reasons(but not limited to these):
- Async communcation (non blocking vs sync blocking)
- Each microservice will be responsible for doing one thing only (ideally)
- Eliminating single points of failures (for api/job services in this case)
- Better isolation
- Easier to scale
- Smaller/Faster/Isolated deployments (better with orchestration, ex. Kubernetes)

**Api server** creates a job, from an object_id, and publishes a "job:created" event with a status of "processing". **Job server** listens for that event, and processes the job (sleeps for random time between 15-45). After being asleep, **job server** will try to publish one of the two possible cases: *cancelled* or *finished*. If the whole operation takes more than 46 seconds (default configured timeout), a "job:cancelled" event will be published, otherwise a "job:finished". **Api server** listens for those types of events, and updates the status accordingly.

- I used nginx as a reverse proxy for making it easier to scale out the components.
- I used NATS Streaming Server for handling the events. Besides being very fast and lightweight, it also resends the message if it's not acknowledged (manually) in a timespan of 50 seconds (service frozen/crashed). I set up a queue group in order to subscribe more consumers to the same channel and only one consumer to receive the message (per queue group). Also, if a new service will become available(in the same queue group), all historical messages will be processed first, in order to be up to date with the rest of the services.
*(Nats Streaming Server gets deprecated, but still receives critical and security fixes - I still chose it for this project, because I'm not yet familiar with the newer versions like JetStream, etc.)*
- Used two mongo dbs for each service

## Diagram
![alt text](https://github.com/bogdan-copocean/hasty-server/raw/main/hasty-server-diagram.png?raw=true)

# Hasty - microservices

**Flow**
- Create job by making a POST request to ```http://localhost/``` with ```{"object_id": "random-object-id"}``` and receives back a job_id
- Check its status at ```http://localhost/job_id```
- Wait 5 minutes before rerunning the job with the same object id (otherwise will get an error)
- If the job processing service goes down, the job will rerun when it comes back up

You can configure a timeout period to cancel the job if needed - **default is 46 seconds**

It can be scaled horizontally by using ```docker compose up --scale service_name=3```

## Installation
I've built the images and pushed them to my docker hub repository, because when running the tests, it actually useses the same docker-compose file when building the environment, and I don't want to build my images every time I'm working on the tests (it takes too much time).

```bash
docker-compose up
```
**or**

*uncomment **context** and **dockerfile** from docker-compose on **api-server and job-server**, and comment image*

```bash
docker-compose up --build
```

## E2E Testing

```go
go test tests/e2e/api_test.go -timeout 200s -v
```
**Test Flow**
- User creates a job with an object id (expects 201 and a job id)
- User tries to create another job with the same object id in less than 5 minutes(expects 400 and 5 minute error)
- User makes a get request with a non existing job id (expects 400 and a no job id *todo: 404 error instead 400*)
- User makes a get request with the received job id (expects 200 and metadata)
- User waits for maximum 45 seconds, and makes the same get request (expects 200, and status updated)
