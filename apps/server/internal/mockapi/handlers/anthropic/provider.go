package anthropic

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
)

// TODO: this is a temporary fixture for the messages endpoint.
//go:embed testdata/analyze_basic.json
var messagesFixture []byte

// TODO: add proper file metadata and content endpoints
func HandleMessages(c *gin.Context) {
	response := map[string]any{}
	err := json.Unmarshal(messagesFixture, &response)
	if err != nil {
		fmt.Println("error unmarshalling messages fixture: ", err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, response)
}

func HandleFileMetadata(c *gin.Context) {
	c.JSON(404, gin.H{"error": "file not found: " + c.Param("id")})
}

func HandleFileContent(c *gin.Context) {
	c.JSON(404, gin.H{"error": "file not found: " + c.Param("id")})
}
