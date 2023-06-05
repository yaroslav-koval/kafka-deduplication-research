package main

import (
	"context"
	"kafka-polygon/pkg/log"
)

func main() {
	ctx := context.Background()

	countOfMessages, err := writeMessages(ctx, "polygon3", 1000000)
	if err != nil {
		log.Info(ctx, err.Error())
		return
	}

	log.InfoF(ctx, "Created %d messages", countOfMessages)
}
