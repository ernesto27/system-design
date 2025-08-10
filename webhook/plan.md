# Webhook System Design

## 1. Requirements

### Functional Requirements
- The system must receive events from external webhooks. 

### Non-Functional Requirements

#### Scale
- **Event Volume**: The system should handle 1 million events per day.
- **Traffic Spikes**: During peak hours, incoming requests may increase by 5 times.
- **Latency Requirement**: End-to-end latency (from event arrival to processing completion) should be under 200 milliseconds.
- **Data Retention**: The system should store all events for 30 days. Assume each event is 5KB.

#### Availability
- **High availability**: The system should be highly available and resilient to failures.

#### Reliability
- **At-least-once processing**: Each event should be processed at least once if the system accepts it.

## 2. API Design

### API Endpoints
`POST /webhook`

- **Description**: Receive webhooks from external systems, returns 200 OK if the event is accepted.
- **Response Body**:
```json
{
  "status": "success"
}
```

Use golang as a backend 
https://echo.labstack.com/docs



## 3. System Architecture
![Sequence Diagram](https://systemdesignschool.io/solutions/webhook/webhook-sequence-diagram.png)

### Event Flow
1. **Send Event**: The external client service (e.g., Shopify.com) triggers an event and sends it to the webhook service's endpoint (our_domain.com/webhook). This event could represent a specific action, like a payment confirmation or order update.
2. **Enqueue Event**: The Request Handler in the webhook service receives the event and enqueues it into a Message Queue. This action stores the event temporarily, allowing the system to process events asynchronously, improving reliability and scalability.
3. **Return 200**: After enqueuing the event successfully, the Request Handler immediately returns a 200 HTTP status code to the client, confirming that the webhook event has been received. This acknowledgment allows the client to know the event was accepted, even if processing hasn't yet occurred.
4. **Fetch Event**: A Queue Consumer fetches the event from the Message Queue. This component is responsible for processing events one by one (or in batches, depending on design) as they become available in the queue.
5. **Process Event**: The Queue Consumer processes the event, which involves performing the necessary operations related to the event. For example, if it’s a payment confirmation, it may update the payment status in the system.
6. **Persist Results**: After processing, the Queue Consumer persists the results of the operation to a database. This could involve storing details like the original event, the outcome of the processing, and any relevant status updates.
7. **DB Write Succeeds**: Once the results are successfully saved in the database, the system receives confirmation of a successful write operation. This ensures that the event processing has been completed and recorded for future reference, such as for audits or debugging.
8. **Dequeue the Event**: After the event has been successfully processed and stored, it is dequeued from the Message Queue. This marks the event as fully handled, removing it from the queue and freeing up space for new incoming events.


Use rabbitmq as a message queue or amazon SQS.
use interface to support both.

start with rabbitmq, then add support for SQS later.


Database: Postgresql or Aws RDS.

Servicios

API 
CONSUMER

DOCKER 
RABBITMQ 
POSTGRESQL


## TODO

Phase 1

- [X] Implement the API endpoint using Golang and Echo framework.
- [X] Set up RabbitMQ with docker compose.
- [X] Set up PostgreSQL with docker compose.
- [X] Create the Queue Consumer to process events from the Message Queue.
- [ ] Create consumer service  
- [] Save processed events to PostgreSQL. 

Phase 2

- [ ] Add support for Amazon SQS as an alternative Message Queue.
- [ ] Implement retry logic for failed event processing.
- [ ] Timeout API requests to ensure they do not hang indefinitely.