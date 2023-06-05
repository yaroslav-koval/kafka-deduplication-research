create-topic-polygon1:
	kafka-topics --create \
    --bootstrap-server localhost:9093 \
    --replication-factor 1 \
    --partitions 1 \
    --topic polygon1

delete-topic-polygon1:
	kafka-topics --bootstrap-server localhost:9093 --delete --topic polygon1

recreate-topic-polygon1:
	make delete-topic-polygon1
	make create-topic-polygon1

create-topic-polygon2:
	kafka-topics --create \
    --bootstrap-server localhost:9093 \
    --replication-factor 1 \
    --partitions 5 \
    --topic polygon2

delete-topic-polygon2:
	kafka-topics --bootstrap-server localhost:9093 --delete --topic polygon2

recreate-topic-polygon2:
	make delete-topic-polygon2
	make create-topic-polygon2

create-topic-polygon3:
	kafka-topics --create \
    --bootstrap-server localhost:9093 \
    --replication-factor 2 \
    --partitions 5 \
    --topic polygon3

delete-topic-polygon3:
	kafka-topics --bootstrap-server localhost:9093 --delete --topic polygon3

recreate-topic-polygon3:
	make delete-topic-polygon3
	make create-topic-polygon3

describe-topic:
	kafka-topics \
	--describe \
	--bootstrap-server localhost:9093 \
	--topic polygon3
