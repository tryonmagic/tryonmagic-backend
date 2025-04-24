package socket

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gofiber/contrib/socketio"
)

// Global variable to store user connections
var (
	Clients      = make(map[string]string) // Map of userId to connection UUID
	clientsMutex = sync.RWMutex{}          // Mutex for thread-safe access
)

// Example functions to interact with the map
func AddClient(userId, connectionUUID string) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()
	Clients[userId] = connectionUUID
}

func RemoveClient(userId string) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()
	delete(Clients, userId)
}

func GetClient(userId string) (string, bool) {
	clientsMutex.RLock()
	defer clientsMutex.RUnlock()
	connectionUUID, exists := Clients[userId]
	return connectionUUID, exists
}

func EmitEventToUser(userId, event string, data any) error {
	fmt.Printf("Client map: %+v\n", Clients)
	fmt.Printf("Emitting event to user: %s, event: %s, data: %+v\n", userId, event, data)
	if connectionUUID, exists := GetClient(userId); exists {
		message := map[string]any{
			"type": event,
			"data": data,
		}
		messageBytes, err := json.Marshal(message)
		if err != nil {
			fmt.Printf("Failed to marshal message: %+v", err)
			return err
		}

		// Emit the event to the specific user
		socketio.EmitTo(connectionUUID, messageBytes)
		fmt.Printf("Notification sent to user: %s\n", userId)
	} else {
		fmt.Printf("No active connection for user: %s\n", userId)
		return fmt.Errorf("no active connection for user: %s", userId)
	}
	return nil
}
