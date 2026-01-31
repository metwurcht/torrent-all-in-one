package gui

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// GUIReporter envoie les événements de progression au frontend
type GUIReporter struct {
	ctx context.Context
}

// NewGUIReporter crée un nouveau reporter pour le GUI
func NewGUIReporter(ctx context.Context) *GUIReporter {
	return &GUIReporter{ctx: ctx}
}

func (g *GUIReporter) OnStart(message string) {
	runtime.EventsEmit(g.ctx, "progress:start", message)
}

func (g *GUIReporter) OnProgress(message string) {
	runtime.EventsEmit(g.ctx, "progress:update", message)
}

func (g *GUIReporter) OnComplete(message string) {
	runtime.EventsEmit(g.ctx, "progress:complete", message)
}

func (g *GUIReporter) OnError(message string) {
	runtime.EventsEmit(g.ctx, "progress:error", message)
}
