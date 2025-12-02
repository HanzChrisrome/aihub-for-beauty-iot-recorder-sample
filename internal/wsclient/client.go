package wsclient

import (
	"encoding/json"
	"log"
	"net/url"
	"time"

	//External Packages:
	"github.com/gorilla/websocket"

	//Internal Packages:
	"github.com/otis-co-ltd/aihub-recorder/internal/config"
)

const (
    MSG_START_RECORDING = "start_recording"
    MSG_STOP_RECORDING  = "stop_recording"
    MSG_LIST_DEVICES    = "list_devices"
    MSG_STOP_ALL        = "stop_all"
    MSG_STATUS          = "status"
    MSG_ERROR           = "error"
    MSG_SUCCESS         = "success"
)

type WSMessage struct {
	Type    string          `json:"type"`
	Data    json.RawMessage `json:"data,omitempty"`
	Message string          `json:"message,omitempty"`
}

type Client struct {
	conn      *websocket.Conn
	send      chan []byte
	piID      string
	serverURL string
}

func Start(piID string) {
	for {
		client, err := connect(piID)
		if err != nil {
			log.Println("[WS] Connection failed:", err)
			time.Sleep(time.Duration(config.ReconnectSeconds) * time.Second)
			continue
		}

		log.Println("[WS] Connected to:", client.serverURL)
		go client.writePump()
		client.readPump()

		log.Println("[WS] Disconnected. Reconnecting...")
		time.Sleep(time.Duration(config.ReconnectSeconds) * time.Second)
	}
}

func connect(piID string) (*Client, error) {
	wsURL := url.URL{
		Scheme: "ws",
		Host:   config.BackendHost,
		Path:   config.WebSocketPath,
	}

	q := wsURL.Query()
	q.Set("pi_id", piID)
	wsURL.RawQuery = q.Encode()

	log.Println("[WS] Dialing:", wsURL.String())

	conn, _, err := websocket.DefaultDialer.Dial(wsURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return &Client{
		conn:      conn,
		send:      make(chan []byte, 256),
		piID:      piID,
		serverURL: wsURL.String(),
	}, nil
}

func (c *Client) readPump() {
	defer func() {
		c.conn.Close()
	}()

	for {
		_, messageBytes, err := c.conn.ReadMessage()
		if err != nil {
			log.Println("? WS read error:", err)
			break
		}

		log.Println("<< Raw WS message:", string(messageBytes))

		var msg WSMessage
		if err := json.Unmarshal(messageBytes, &msg); err != nil {
			log.Println("? WS invalid JSON, treating as status:", string(messageBytes))
			c.handleMessage(WSMessage{
				Type:    MSG_STATUS,
				Message: string(messageBytes),
			})
			continue
		}

		c.handleMessage(msg)
	}
}

func (c *Client) writePump() {
	defer c.conn.Close()

	for message := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Println("? WS write error:", err)
			return
		}
	}
}
