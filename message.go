package hooksink

import (
	"encoding/json"
	"errors"

	"github.com/slack-go/slack"
)

// Message is the stable top-level representation of a Slack-compatible
// incoming webhook payload.
//
// Nested Block Kit and attachment structures use slack-go types so this package
// can track Slack schema changes without hand-rolling those trees. Extra holds
// unknown top-level JSON fields so forward-compatible payload data is not lost.
type Message struct {
	Text         string                     `json:"text,omitempty"`
	Blocks       slack.Blocks               `json:"blocks,omitempty"`
	Attachments  []slack.Attachment         `json:"attachments,omitempty"`
	Username     string                     `json:"username,omitempty"`
	IconEmoji    string                     `json:"icon_emoji,omitempty"`
	IconURL      string                     `json:"icon_url,omitempty"`
	Channel      string                     `json:"channel,omitempty"`
	Mrkdwn       *bool                      `json:"mrkdwn,omitempty"`
	ResponseType string                     `json:"response_type,omitempty"`
	Extra        map[string]json.RawMessage `json:"-"`
}

type messageWire struct {
	Text         string             `json:"text,omitempty"`
	Blocks       slack.Blocks       `json:"blocks,omitempty"`
	Attachments  []slack.Attachment `json:"attachments,omitempty"`
	Username     string             `json:"username,omitempty"`
	IconEmoji    string             `json:"icon_emoji,omitempty"`
	IconURL      string             `json:"icon_url,omitempty"`
	Channel      string             `json:"channel,omitempty"`
	Mrkdwn       *bool              `json:"mrkdwn,omitempty"`
	ResponseType string             `json:"response_type,omitempty"`
}

var messageJSONFields = map[string]struct{}{
	"text":          {},
	"blocks":        {},
	"attachments":   {},
	"username":      {},
	"icon_emoji":    {},
	"icon_url":      {},
	"channel":       {},
	"mrkdwn":        {},
	"response_type": {},
}

// UnmarshalJSON decodes a Message and preserves unknown top-level fields in
// Extra.
func (m *Message) UnmarshalJSON(data []byte) error {
	var all map[string]json.RawMessage
	if err := json.Unmarshal(data, &all); err != nil {
		return err
	}
	if all == nil {
		return errors.New("hooksink: message payload must be a JSON object")
	}

	var wire messageWire
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}

	extra := make(map[string]json.RawMessage)
	for key, raw := range all {
		if _, known := messageJSONFields[key]; known {
			continue
		}
		extra[key] = raw
	}

	*m = Message{
		Text:         wire.Text,
		Blocks:       wire.Blocks,
		Attachments:  wire.Attachments,
		Username:     wire.Username,
		IconEmoji:    wire.IconEmoji,
		IconURL:      wire.IconURL,
		Channel:      wire.Channel,
		Mrkdwn:       wire.Mrkdwn,
		ResponseType: wire.ResponseType,
	}
	if len(extra) > 0 {
		m.Extra = extra
	}

	return nil
}

// MarshalJSON encodes a Message, including unknown top-level fields from Extra.
func (m Message) MarshalJSON() ([]byte, error) {
	fields := make(map[string]json.RawMessage, len(m.Extra)+len(messageJSONFields))
	for key, raw := range m.Extra {
		if _, known := messageJSONFields[key]; known {
			continue
		}
		if raw == nil {
			continue
		}
		fields[key] = append(json.RawMessage(nil), raw...)
	}

	if m.Text != "" {
		if err := putJSONField(fields, "text", m.Text); err != nil {
			return nil, err
		}
	}
	if len(m.Blocks.BlockSet) > 0 {
		if err := putJSONField(fields, "blocks", m.Blocks); err != nil {
			return nil, err
		}
	}
	if len(m.Attachments) > 0 {
		if err := putJSONField(fields, "attachments", m.Attachments); err != nil {
			return nil, err
		}
	}
	if m.Username != "" {
		if err := putJSONField(fields, "username", m.Username); err != nil {
			return nil, err
		}
	}
	if m.IconEmoji != "" {
		if err := putJSONField(fields, "icon_emoji", m.IconEmoji); err != nil {
			return nil, err
		}
	}
	if m.IconURL != "" {
		if err := putJSONField(fields, "icon_url", m.IconURL); err != nil {
			return nil, err
		}
	}
	if m.Channel != "" {
		if err := putJSONField(fields, "channel", m.Channel); err != nil {
			return nil, err
		}
	}
	if m.Mrkdwn != nil {
		if err := putJSONField(fields, "mrkdwn", *m.Mrkdwn); err != nil {
			return nil, err
		}
	}
	if m.ResponseType != "" {
		if err := putJSONField(fields, "response_type", m.ResponseType); err != nil {
			return nil, err
		}
	}

	return json.Marshal(fields)
}

func putJSONField(fields map[string]json.RawMessage, key string, value any) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	fields[key] = raw
	return nil
}
