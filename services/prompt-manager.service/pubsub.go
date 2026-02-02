package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"cloud.google.com/go/pubsub"
)

const (
	topicID        = "prompt-requests"
	subscriptionID = "prompt-worker-sub"
)

var (
	pubsubClient *pubsub.Client
	topic        *pubsub.Topic
)

func initPubSub() error {
	ctx := context.Background()

	// Get project ID from environment or use default for emulator
	projectID := os.Getenv("PUBSUB_PROJECT_ID")
	if projectID == "" {
		projectID = "local-project"
	}

	var err error
	pubsubClient, err = pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Printf("Failed to create Pub/Sub client: %v", err)
		return err
	}

	// Create topic if it doesn't exist
	topic = pubsubClient.Topic(topicID)
	exists, err := topic.Exists(ctx)
	if err != nil {
		log.Printf("Error checking topic existence: %v", err)
		return err
	}

	if !exists {
		topic, err = pubsubClient.CreateTopic(ctx, topicID)
		if err != nil {
			log.Printf("Error creating topic: %v", err)
			return err
		}
		log.Printf("Created topic: %s", topicID)
	}

	// Create subscription if it doesn't exist
	sub := pubsubClient.Subscription(subscriptionID)
	exists, err = sub.Exists(ctx)
	if err != nil {
		log.Printf("Error checking subscription existence: %v", err)
		return err
	}

	if !exists {
		_, err = pubsubClient.CreateSubscription(ctx, subscriptionID, pubsub.SubscriptionConfig{
			Topic:       topic,
			AckDeadline: 60 * time.Second,
		})
		if err != nil {
			log.Printf("Error creating subscription: %v", err)
			return err
		}
		log.Printf("Created subscription: %s", subscriptionID)
	}

	log.Println("Pub/Sub initialized successfully")
	return nil
}

func publishMessage(msg PubSubMessage) error {
	ctx := context.Background()

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return err
	}

	result := topic.Publish(ctx, &pubsub.Message{
		Data: data,
	})

	// Block until the result is returned
	id, err := result.Get(ctx)
	if err != nil {
		log.Printf("Error publishing message: %v", err)
		return err
	}

	log.Printf("Published Pub/Sub message ID: %s for chat: %s", id, msg.ChatID)
	return nil
}

func closePubSub() {
	if topic != nil {
		topic.Stop()
	}
	if pubsubClient != nil {
		pubsubClient.Close()
	}
}
