package wsclient

import (
	"encoding/json"
	"log"
)

// sendSuccessMessage sends a success response to the server
func (c *Client) sendSuccessMessage(command, message string) {
    response := ResponseMessage{
        Command: command + "_response",
        Status:  "success",
        Message: message,
    }
    c.sendResponse(response)
}

// sendErrorMessage sends an error response to the server
func (c *Client) sendErrorMessage(command, message string) {
    response := ResponseMessage{
        Command: command + "_response",
        Status:  "error",
        Message: message,
    }
    c.sendResponse(response)
}

// sendResponse sends a structured response message
func (c *Client) sendResponse(response ResponseMessage) {
    data, err := json.Marshal(response)
    if err != nil {
        log.Printf("Failed to marshal response: %v", err)
        return
    }

    c.send <- data
}

// Backward compatibility functions
func (c *Client) sendError(errMsg string) {
    msg := WSMessage{
        Type:    MSG_ERROR,
        Message: errMsg,
    }
    
    data, err := json.Marshal(msg)
    if err != nil {
        log.Printf("Failed to marshal error message: %v", err)
        return
    }
    
    c.send <- data
}

func (c *Client) sendSuccess(successMsg string) {
    msg := WSMessage{
        Type:    MSG_SUCCESS,
        Message: successMsg,
    }
    
    data, err := json.Marshal(msg)
    if err != nil {
        log.Printf("Failed to marshal success message: %v", err)
        return
    }
    
    c.send <- data
}