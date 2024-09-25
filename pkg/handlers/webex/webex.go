package webex

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"

	webex "github.com/jbogarin/go-cisco-webex-teams/sdk"

	"github.com/bitnami-labs/kubewatch/config"
	"github.com/bitnami-labs/kubewatch/pkg/event"
)

var webexErrMsg = `
%s

You need to set both webex token and room for webex notify,
using "--token/-t", "--room/-r", and "--url/-u" or using environment variables:

export WEBEX_ACCESS_TOKEN=webex_token
export WEBEX_ACCESS_ROOM=webex_room
export WEBEX_NOTIFICATION_LABEL=k8sClusterName
export WEBEX_SKIP_NAMESPACES=NamespaceToSkip,AnotherNamespaceToSkip

Command line flags will override environment variables

`

// Webex handler implements handler.Handler interface,
// Notify event to Webex room
type Webex struct {
	Token 				string
	Room  				string
	NotificationLabel   string
	SkipNamespaces      string
}

// Init prepares Webex configuration
func (s *Webex) Init(c *config.Config) error {
	notificationlabel := c.Handler.Webex.NotificationLabel
	room := c.Handler.Webex.Room
	token := c.Handler.Webex.Token
	skipnamespaces := c.Handler.Webex.SkipNamespaces

	if token == "" {
		token = os.Getenv("WEBEX_ACCESS_TOKEN")
	}

	if skipnamespaces == "" {
		skipnamespaces = os.Getenv("WEBEX_SKIP_NAMESPACES")
	}

	if room == "" {
		room = os.Getenv("WEBEX_ROOM")
	}

	if notificationlabel == "" {
        notificationlabel = os.Getenv("WEBEX_NOTIFICATION_LABEL")
	}

	s.Token = token
	s.Room = room
	s.NotificationLabel = notificationlabel
	s.SkipNamespaces = skipnamespaces

	return checkMissingWebexVars(s)
}

// Handle handles the notification.
func (s *Webex) Handle(e event.Event) {
	if !contains(strings.Split(s.SkipNamespaces, ","), e.Namespace){
		client := webex.NewClient()
		client.SetAuthToken(s.Token)
		message := &webex.MessageCreateRequest{
			RoomID: s.Room,
			Text:   "From " + s.NotificationLabel + ": " + e.Message(),
		}
		_, response, err := client.Messages.CreateMessage(message)
		if err != nil {
			fmt.Println("Error sending message:", err)
			return
		}
		logrus.Printf("Message sent: Return Code %d", response.StatusCode())
		logrus.Printf("Message successfully sent to room %s", s.Room)
	} else {
		logrus.Printf("Message skipped for namespace: %s", e.Namespace)
	}

}

func checkMissingWebexVars(s *Webex) error {
	if s.Token == "" || s.Room == "" {
		return fmt.Errorf(webexErrMsg, "Missing webex token or room")
	}

	return nil
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if os.Getenv("DEBUG") == "true" {
			fmt.Printf("check %s = %s \n", str, v)
		}
		if v == str {
			return true
		}
	}
	return false
}