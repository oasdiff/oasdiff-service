package internal

import "github.com/oasdiff/telemetry/client"

type Handler struct{ collector *client.Collector }

func NewHandler(c *client.Collector) *Handler {
	return &Handler{collector: c}
}
