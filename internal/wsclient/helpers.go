package wsclient

import (
	"encoding/json"
	"log"
)

func (c *Client) sendMessage(msg WSMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Println("? WS marshal error:", err)
		return
	}

	c.send <- data
}

func (c *Client) sendSuccess(message string) {
	c.sendMessage(WSMessage{
		Type:    MSG_SUCCESS,
		Message: message,
	})
}

func (c *Client) sendError(message string) {
	c.sendMessage(WSMessage{
		Type:    MSG_ERROR,
		Message: message,
	})
}
