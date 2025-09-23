//=============================================================================
/*
Copyright © 2025 Andrea Carboni andrea.carboni71@gmail.com

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
//=============================================================================

package messaging

import (
	"encoding/json"
	"html/template"
	"log/slog"
	"strings"
	"time"

	"github.com/bit-fever/core/msg"
	"github.com/bit-fever/event-store/pkg/db"
	"gorm.io/gorm"
)

//=============================================================================

type SentEvent struct {
	Username   string
	Level      msg.EventLevel
	EventDate  time.Time
	Code       string
	Title      string
	Message    string
	Parameters map[string]any
}

//=============================================================================

func InitMessageListener() {
	slog.Info("Starting message listeners...")
	initTemplates()

	go msg.ReceiveMessages(msg.QuAllToEvent,  handleEventMessage)
}

//=============================================================================

func handleEventMessage(m *msg.Message) bool {

	slog.Info("handleEventMessage: New event message received", "source", m.Source, "type", m.Type)

	if m.Source == msg.SourceEvent {
		ev  := SentEvent{}
		err := json.Unmarshal(m.Entity, &ev)
		if err != nil {
			slog.Error("handleEventMessage: Dropping badly formatted message for event!", "entity", string(m.Entity))
			return true
		}

		if m.Type == msg.TypeCreate {
			return handleNewEvent(&ev)
		}
	}

	slog.Error("handleEventMessage: Dropping message with unknown source/type!", "source", m.Source, "type", m.Type)
	return true
}

//=============================================================================

func handleNewEvent(ev *SentEvent) bool {
	title, message, level := ev.Title, ev.Message, ev.Level

	if ev.Code != "" {
		title, message, level = getTitleAndMessage(ev.Code)
	}

	title, message = fillInParameters(title, message, ev.Parameters)

	params, err := json.Marshal(ev.Parameters)
	if err != nil {
		slog.Error("handleNewEvent: Error marshalling parameters", "error", err.Error())
		return false
	}

	if len(title) > 64 {
		slog.Warn("handleNewEvent: Title is too long. Clipping", "title", title)
		title = title[0:64]
	}

	if len(message) > 512 {
		slog.Warn("handleNewEvent: Message is too long. Clipping", "message", message)
		message = message[0:512]
	}

	//--- Store event

	dvE := db.Event{
		Username:   ev.Username,
		EventDate:  ev.EventDate,
		Level:      level,
		Title:      title,
		Message:    message,
		Parameters: params,
	}

	err = db.RunInTransaction(func(tx *gorm.DB) error {
		return db.AddEvent(tx, &dvE)
	})

	if err != nil {
		slog.Error("handleNewEvent: Error adding event to the database", "error", err.Error())
	}

	return err == nil
}

//=============================================================================

func getTitleAndMessage(code string) (string, string, msg.EventLevel) {
	t,ok := templates[code]
	if ok {
		return t.Title, t.Message, convertLevel(t.Level)
	}

	return "?"+code+"?", "?Code not found?", msg.EventLevelError
}

//=============================================================================

func convertLevel(level string) msg.EventLevel {
	switch level {
		case "INFO": return msg.EventLevelInfo
		case "WARN": return msg.EventLevelWarning
		default    : return msg.EventLevelError
	}
}

//=============================================================================

func fillInParameters(title, message string, parameters map[string]any) (string, string) {
	var err error

	title,err = fillTemplate(title, parameters)
	if err != nil {
		return err.Error(), message
	}

	message,err = fillTemplate(message, parameters)
	if err != nil {
		return title, err.Error()
	}

	return title, message
}

//=============================================================================

func fillTemplate(temp string, parameters map[string]any) (string, error) {
	var err error

	t := template.New("tmp")
	t, err = t.Parse(temp)

	if err != nil {
		return "", err
	}

	w := strings.Builder{}
	err = t.Execute(&w, parameters)
	if err != nil {
		return "", err
	}

	return w.String(), nil
}

//=============================================================================
