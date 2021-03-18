# hobbyq: a hobby-grade message bus

hobbyq aims to fill the same niche as RabbitMQ and Amazon's SNS/SQS
combination, but in a very lightweight way. The most important
limitation is that it is not distributed: it's a single-process,
single-node message bus.

Here are the features I'm aiming to provide:

* exchanges for fanout (just like RabbitMQ)
* durable or temporary queues, at the consumer's choice (like RabbitMQ)
* messages are pushed to consumers over a socket
* or messages can be pushed by a webhook
