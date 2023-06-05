# Kafka polygon

### Before getting started

- You can use [deployments](deployments) to deploy one or several brokers kafka infrastructure
- Makefile uses **kafka-topics** to create, delete, describe topics. It can be installed with **brew install kafka**

### It's a repository where are presented 3 most used kafka-go libraries.

- [segmentio/kafka-go](https://github.com/segmentio/kafka-go)
- [confluent-kafka-go](https://github.com/confluentinc/confluent-kafka-go)
- [shopify/sarama](https://github.com/Shopify/sarama)

*Wasfaty* uses **segmentio** library, buy it has lack of capabilities, like absence of *idempotent* or *transactional*
producers, inability to set up kafka precisely and a bug that leads to records duplication during brokers rebalancing.

#### Idempotence in segmentio/kafka-go

I didn't find any mention of idempotence in release notes. Also, I
found [this](https://github.com/segmentio/kafka-go/issues/797#issuecomment-1165706820)
and [this](https://github.com/redpanda-data/redpanda/issues/7835#issuecomment-1359038821) GitHub issue conversations.
Even if library support it, it can't be tested or hard to
test. [Small hint why
](https://stackoverflow.com/questions/75569262/why-i-still-got-duplicate-messages-in-kafka-after-i-set-enable-idempotence-tru)
on stackoverflow. Also, [this video](https://youtu.be/B8glj1BJkSw?t=272) can provide more deep information.

#### Transactions in segmentio/kafka-go

Also **segmentio/kafka-go** doesn't have transactions:

1. Writer has [**produce** function
   ](https://github.com/segmentio/kafka-go/blob/f057b1d369cdb68b249e6a9b12f0d797ebfae407/writer.go#LL730C34-L730C34)
   where it calls client's Produce, but doesn't pass *TransactionalID*. However, client can receive *TransactionalID* in
   config.
2. In GitHub exists [conversation](https://github.com/segmentio/kafka-go/issues/897) about usage in transactions with
   link to draft [PR](https://github.com/segmentio/kafka-go/pull/755) from 06.10.2021

### Bug with records duplication:

If you start [segmentio consumer](segmentio/consumer/main.go) you'll find out that almost after each rebalancing can be
found
message starting with **Handled message received** in [logs](segmentio/consumer/logs) folder. But if we run program, we
can see more information about rebalancing in console. It's important to keep several consumers not closed, so they can
rebalance after new consumer added.

The most annoying thing is that **segmentio** library doesn't have enough config settings to change behavior of
rebalancing. So was made decision to migrate our project to another library and candidates are **confluent** and
**sarama** libraries.

*Summary*, we have 3 problems:

1. Idempotent producer
2. Transactional producer
3. Message duplication on kafka rebalancing

### Confluent

After a little research **confluent** has been chosen as reliable library with core written with C and called
[librdkafka](https://github.com/confluentinc/librdkafka). And **confluent** uses CGo to call C code inside, it can slow
system. So this library has a lot more capabilities in producer/consumer set up by untyped config fields, what makes it
more hard in use but extend ability of configuring to native kafka level. And it's able to create **idempotent**
**transactional** producer, that already covers 2 of 3 problems. Configs creation:

<table>
<thead><tr><th>Segmentio producer</th> <th>Confluent producer</th></tr></thead>
<tbody>
<tr>
<td>

```go
d, _ := time.ParseDuration("5s")
d2, _ := time.ParseDuration("50ms")
d3, _ := time.ParseDuration("3s")

return &kafka.Config{
Brokers:                []string{"localhost:9093"},
LoggerEnabled:          true,
UseKeyDoubleQuote:      true,
AllowAutoTopicCreation: false,
Producer: kafka.Producer{
MaxAttempts:        3,
MaxRetry:           10,
MaxAttemptsDelay:   d,
WriteTimeout:       d2,
WriterBatchSize:    100,
WriterBatchTimeout: d3,
},
}

```

</td>
<td>

```go
config := kafka.ConfigMap{
"bootstrap.servers":                     broker,
"enable.idempotence":                    true,
"acks":                                  "all",
"retries":                               10,
"retry.backoff.ms":                      500,
"delivery.timeout.ms":                   3000,
"max.in.flight.requests.per.connection": 5,
}
```

</td>
</tr>

</tbody>
</table>

##### Some configs are stored in consumer folders in *configs.go* files for each library.

[For segmentio](segmentio/consumer/configs.go). [For confluent](confluent/consumer/configs.go). [For sarama](sarama/consumer/configs.go).

After tests this library showed good results and bug with rebalancing seemed to go away. Third problem is resolved. But
after adding more than 1 partition in topic (exactly 5) just with 1 broker duplicate messages appeared again. Only one
library remained - **sarama**

### Sarama

For now **sarama** has most used repo in GitHub compared to previous two. It combines advantages of configuring
producers/consumers using typed config structure in one hand and a lot of different settings in other hand. After tests
**sarama** showed it capability to create **idempotent** and **transactional** producers, also it has no problems with
rebalancing during consuming. It's showed durability under load, but also created one new problem.

**Sarama** uses context to cancel message consuming and **pkg** uses [go-redis](https://github.com/redis/go-redis)
which also needs context to get or delete keys. **pkg** has broker logic to validate message on duplication.
**pkg/broker** firstly gets information about message and, if nothing found, pushes message to redis as consumed, before
business logic is called. To get message is used function **GetEventInfoByID**, to put message is used
**PutEventInfoByID**. Trouble starts at the moment when **GetEventInfoByID** gets nothing, context is getting cancelled
between **GetEventInfoByID** and **PutEventInfoByID** functions and then **PutEventInfoByID** is called. **go-redis**
returns error after **PutEventInfoByID** function, but it also pushes record to redis. For now, we can either change
redis library or delete message from redis manually after checking that error was exactly 'context canceled' as it's
done in [**ConsumeClaim** method in sarama consumer's handler](sarama/consumer/handler.go).
Reproduction:

1. Comment out **DeleteEventInfoByID**
2. Try to restart program several times until we get *context cancelled* error from **PutEventInfoByID** method, not
   **GetEventInfoByID**
3. Next time you run program first message will be duplication

**Little deviation**: library needs *PanicHandler* function realization

## Research remarks

### Common rebalancing issue

We can have 2 "new" status messages from redis like in **sarama**'s example above, not only because of context
cancellation. We can get second message if rebalancing happens in time between first consumer calls **GetEventInfoByID**
and **f.CallFn(ctx, e, eventData)** in [**Run** function](pkg/broker/provider/provider.go). Sarama covers this issue in
general by using rebalance strategy **sticky**.

## Useful links

### Confluent's set of articles about exactly-one delivery in kafka

1. [Exactly-Once Semantics Are Possible: Here’s How Kafka Does It
   ](https://www.confluent.io/blog/exactly-once-semantics-are-possible-heres-how-apache-kafka-does-it/)
2. [Transactions in Apache Kafka
   ](https://www.confluent.io/blog/transactions-apache-kafka/)
3. [Enabling Exactly-Once in Kafka Streams
   ](https://www.confluent.io/blog/enabling-exactly-once-kafka-streams/)

### Common info

- [Exactly Once Delivery](https://docs.google.com/document/d/11Jqy_GjUGtdXJK94XGsEIK7CP1SnQGdp2eF0wSw9ra8/edit#)
  documentation
- [Kafka Partitions and Consumer Groups in 6 mins
  ](https://medium.com/javarevisited/kafka-partitions-and-consumer-groups-in-6-mins-9e0e336c6c00)
