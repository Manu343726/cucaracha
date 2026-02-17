package core

import "github.com/Manu343726/cucaracha/pkg/utils"

type EventHandler interface {
	EventFired(event *Event) bool
}

func NewEventHandler(callback EventCallback) EventHandler {
	return &eventHandlerFromCallback{
		callback: callback,
	}
}

type eventHandlerFromCallback struct {
	callback EventCallback
}

func (h *eventHandlerFromCallback) EventFired(event *Event) bool {
	return h.callback(event)
}

type EventSubscription struct {
	Event   DebugEvent
	Handler EventHandler
	Once    bool
}

// Implements an event handling mechanism for debugger internal purposes
//
// It allows the debugger to subscribe to events and react to them, without
// interfering with the user-defined event callback.
//
// It acts as a queue, allowing multiple event handlers to be registered for the same event.
// When an event is fired, all registered handlers for that event are called in the order they were added.
type InternalEventsQueue struct {
	subscriptors []EventSubscription
}

func NewInternalEventsQueue() *InternalEventsQueue {
	return &InternalEventsQueue{
		subscriptors: make([]EventSubscription, 0),
	}
}

// Adds an event handler for a specific event
func (q *InternalEventsQueue) Subscribe(event DebugEvent, handler EventHandler) {
	q.subscriptors = append(q.subscriptors, EventSubscription{
		Event:   event,
		Handler: handler,
	})
}

// Adds an event handler for a specific event, and ensures the handler is fired only once (the first time the event is fired)
func (q *InternalEventsQueue) SubscribeOnce(event DebugEvent, handler EventHandler) {
	q.subscriptors = append(q.subscriptors, EventSubscription{
		Event:   event,
		Handler: handler,
		Once:    true,
	})
}

// Removes an event handler
func (q *InternalEventsQueue) Unsubscribe(handler EventHandler) {
	q.subscriptors = utils.Filter(q.subscriptors, func(sub EventSubscription) bool {
		return sub.Handler != handler
	})
}

// Removes all event handlers
func (q *InternalEventsQueue) ClearEventHandlers() {
	q.subscriptors = make([]EventSubscription, 0)
}

// Notifies all subscribed handlers of an event
// Returns true if all handlers returned true, false otherwise
func (q *InternalEventsQueue) EventFired(event *Event) bool {
	shouldContinue := true
	subscriptorsToKeep := make([]EventSubscription, 0, len(q.subscriptors))

	for _, subscriptor := range q.subscriptors {
		if subscriptor.Event != event.Event {
			continue
		}

		if !subscriptor.Handler.EventFired(event) {
			shouldContinue = false
		}

		if !subscriptor.Once {
			subscriptorsToKeep = append(subscriptorsToKeep, subscriptor)
		}
	}

	q.subscriptors = subscriptorsToKeep

	return shouldContinue
}
