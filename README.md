# Chainlink Lite

Chainlink Lite is a decentralized application that demonstrates a simple implementation of a gossiping node system. This system, every 30 seconds, pulls and signs the price of Ethereum (ETH) from the CoinGecko API and emits it to a libp2p gossip network. The nodes in the network successively sign the message and re-emit it. When a message has received at least 3 signatures and it has been at least 30 seconds since the last message was written to the database, a node writes this message, including its signatures, to a shared PostgreSQL instance.


## System Workflow

1. Price Fetching and Signing: Each node independently fetches the ETH price and signs it using its ECDSA key pair.
2. Gossip Broadcast: Nodes broadcast their signed messages to the network.
3. Message Reception and Re-signing: Nodes receive messages, verify signatures, and add their own signature before re-broadcasting.
4. Signature Threshold: Once a message accumulates at least 3 signatures, it becomes eligible for database storage.
5. Database Write (Conditional): A node writes the message to the database if 30 seconds have passed since the last write, preventing database flooding.


## Getting Started

### Prerequisites

- Docker and Docker Compose
- Go 1.22 or later


### Installation

1. Clone the repository:

    ```sh
    git clone https://github.com/fernandofcampos/chainlink-lite.git
    cd chainlink-lite
    ```

2. Build the Docker images:
    ```sh
    docker compose build
    ```

3. Start the services:
    ```sh
    docker compose up -d
    ```

4. Stop the services:
    ```sh
    docker compose stop
    ```

### Configuration
The application can be configured via the [config.yaml](config/config.yaml) file. Here you can set the database URL, the price ticker API URL, pubsub topic, and other settings related to the gossiping behavior and logging. Configurations are documented in the file. 


### Logs

- Log level can be set at [config.yaml](config/config.yaml) file. 
- Many events are logged using the `Debug` level. 
- I recommend using [Docker Desktop](https://www.docker.com/products/docker-desktop/) for logs and metrics visualization.


### Adjusting the Number of Nodes

The number of nodes in the system can be adjusted by setting the `replicas` variable in the [docker-compose.yml](docker-compose.yml) file.


### Viewing Database Values
PostgreSQL configuration on Docker Compose exposes port 5432 to localhost. To connect to the PostgreSQL database and view the values being added, you can use any PostgreSQL client with the connection string provided in the config/config.yaml file. For example:
```sh
psql postgres://user:password@localhost:5432/chainlinklite
```

Then, you can query the database table where the price messages are stored:
```sql
SELECT * FROM eth_price_messages;
```

## Design Decisions

### GossipSub:

GossipSub was chosen for the simplified Chainlink oracle system because of its strengths and alignment with our specific requirements:

- Efficient Broadcast: GossipSub is designed for efficient message dissemination in large, decentralized networks. It uses a combination of mesh networking and gossip protocols to ensure that messages reach interested peers quickly and reliably.
- Scalability: GossipSub scales well to large networks, making it suitable for a Chainlink oracle system that may potentially involve many nodes.
- Topic-Based Subscription: GossipSub allows nodes to subscribe to specific topics, ensuring they only receive relevant price updates. This reduces unnecessary network traffic and improves efficiency.
- Robustness: GossipSub is designed to be resilient to network failures and churn. It can handle node disconnections and reconnections gracefully, ensuring continuous data propagation.
- Proven in Production: GossipSub is used in various production systems, including Filecoin and Ethereum 2.0, demonstrating its reliability and effectiveness in real-world scenarios.
- Transport Encryption: uses TLS 1.3 by default for secure communication between peers.
- Security: implements message signing and verification.
- Used the public DHT bootstrap peers provided by libp2p.

### Code
- The code is organized using the clean architecture principles.

### Database
- Schema:

    ```sql
        CREATE TABLE eth_price_messages (
            id SERIAL PRIMARY KEY,
            message_id TEXT NOT NULL UNIQUE,
            price NUMERIC NOT NULL,
            publisher TEXT NOT NULL,
            writer TEXT NOT NULL,
            signers TEXT[] NOT NULL,
            signatures JSONB NOT NULL,
            created_at TIMESTAMPTZ NOT NULL,
            timestamp TIMESTAMPTZ NOT NULL
        );

        CREATE INDEX idx_messages_timestamp ON eth_price_messages (timestamp DESC);
    ```

    - `message_id`: a nounce created when the message is published for the first time.
    - `price`: the price of Ethereum, in decimals.
    - `publisher`: the id of the node that originally published the message.
    - `writer`: the id of the node that wrote the message into the DB.
    - `signers`: list of node ids that signed the message.
    - `signatures`: list of signatures.
    - `created_at`: time when the message was created.
    - `timestamp`: time when the message was inserted into the DB.

- The index on timestamp is used to make the query to find the time of the last write more efficient. Since the number of writes is low (1 every 30 seconds) compared to the number of reads, there's not much overhead in keeping the index.
- To prevent race conditions, I used Postgres advisory locks to create an atomic operation for checking the time of the latest write, and writing the message into the database if at least 30 seconds has passed.


### Mock price ticker

The Coingecko API has a rate limit of 30 requests/min. Depending on the number of nodes you wish to run, the service will start to throttle. 

I created a mock price ticker that returns a random price value. To enable it, set this on [config.yaml](config/config.yaml):

```yml
price_ticker:
  mock: true
```

### Postgres configuration

For simplification, I used environment variables on [docker-compose.yml](docker-compose.yml) and a [init.sql](config/init.sql) file that runs the first time Postgres runs. This configuration is not appropriate for a production environment.


### Docker image optimization

- Use multi-stage build, and use a minimal image for the runtime environment.
- Compress the binary using `upx`.
- Exclude unnecessary parts of the Go standard library or dependencies by using build tags.
- Strip Debug Information.
- Result: 
    - Image size with original Dockerfile: 1.69 GB
    - Image size with optimized Dockerfile: **10.2 MB**


## Future Improvements

- Implement topic-based authorization to control who can publish and subscribe to specific topics. This helps prevent unauthorized actors from injecting messages into the network.

- Implement a peer reputation system. This allows nodes in the network to rate each other based on behavior. Nodes with low reputations can be penalized or excluded, further hardening the network against malicious actors.

- Add a filter to discard prices that deviate significantly from the aggregated value. This can protect against malicious nodes attempting to manipulate the data.

- Implement a more robust authentication mechanism for nodes joining the network. Consider using a public key infrastructure (PKI) or a permissioned blockchain to manage node identities.

- Encrypt the price data being transmitted over the libp2p network to prevent eavesdropping and tampering.

- Store private keys used for signing messages in secure environments, such as hardware security modules (HSMs) or encrypted key stores Implement robust procedures for key generation, rotation, and revocation.

- Implement private bootstrap nodes. 

- Store the public keys of trusted nodes into the system to prevent malicious nodes from joining and injecting false data.

- Integrate with multiple data APIs.

- Secure the database with strong authentication and authorization to prevent unauthorized writes.

- Testing: Add unit and integration tests to ensure the reliability and correctness of the system.

- Use database migrations for database schema management.

- Implement a retry mechanism for network requests and database operations to handle transient failures.

- Use a more sophisticated configuration management system (e.g., Consul, etcd) for dynamic configuration updates.

- Enhance logging with more detailed information and structured logs (e.g., JSON format) for easier debugging and analysis.

- Use dependency injection to improve testability and flexibility.


