// Package hooksink receives Slack-compatible incoming webhook payloads.
//
// This package is the server side of Slack's incoming webhook wire format: it
// parses JSON and legacy form-encoded payloads that third-party tools normally
// POST to hooks.slack.com. It is not a Slack client, does not send messages to
// Slack, and does not mock the Slack Web API.
package hooksink
