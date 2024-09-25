package webex

import (
	"fmt"
	"os"
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

Command line flags will override environment variables

`

// Webex handler implements handler.Handler interface,
// Notify event to Webex room
type Webex struct {
	Token 				string
	Room  				string
	NotificationLabel   string
	Skip 				[]config.SkipItem
}

// Init prepares Webex configuration
func (s *Webex) Init(c *config.Config) error {
	notificationlabel := c.Handler.Webex.NotificationLabel
	room := c.Handler.Webex.Room
	token := c.Handler.Webex.Token
	skip := c.Handler.Webex.Skip

	if token == "" {
		token = os.Getenv("WEBEX_ACCESS_TOKEN")
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
	s.Skip = skip
	return checkMissingWebexVars(s)
}

// Handle handles the notification.
func (s *Webex) Handle(e event.Event) {
	for i := range s.Skip {
		if s.Skip[i].Namespace == e.Namespace && (s.Skip[i].Kind == e.Kind || s.Skip[i].Kind == "*") {
			logrus.Printf("%s messages skipped for namespace: %s",e.Kind, e.Namespace)
            return
        }
    }
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

}

func checkMissingWebexVars(s *Webex) error {
	if s.Token == "" || s.Room == "" {
		return fmt.Errorf(webexErrMsg, "Missing webex token or room")
	}

	return nil
}