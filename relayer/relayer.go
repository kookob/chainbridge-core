// Copyright 2021 ChainSafe Systems
// SPDX-License-Identifier: LGPL-3.0-only

package relayer

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/codes"

	"github.com/ChainSafe/chainbridge-core/relayer/message"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	traceapi "go.opentelemetry.io/otel/trace"
)

type DepositMeter interface {
	TrackDepositMessage(m *message.Message)
	TrackExecutionError(m *message.Message)
	TrackSuccessfulExecutionLatency(m *message.Message)
}

type RelayedChain interface {
	PollEvents(ctx context.Context, sysErr chan<- error, msgChan chan []*message.Message)
	Write(ctx context.Context, messages []*message.Message) error
	DomainID() uint8
}

func NewRelayer(chains []RelayedChain, metrics DepositMeter, messageProcessors ...message.MessageProcessor) *Relayer {
	return &Relayer{relayedChains: chains, messageProcessors: messageProcessors, metrics: metrics}
}

type Relayer struct {
	metrics           DepositMeter
	relayedChains     []RelayedChain
	registry          map[uint8]RelayedChain
	messageProcessors []message.MessageProcessor
}

// Start function starts the relayer. Relayer routine is starting all the chains
// and passing them with a channel that accepts unified cross chain message format
func (r *Relayer) Start(ctx context.Context, sysErr chan error) {
	log.Debug().Msgf("Starting relayer")

	messagesChannel := make(chan []*message.Message)
	for _, c := range r.relayedChains {
		log.Debug().Msgf("Starting chain %v", c.DomainID())
		r.addRelayedChain(c)
		go c.PollEvents(ctx, sysErr, messagesChannel)
	}

	for {
		select {
		case m := <-messagesChannel:
			go r.route(m)
			continue
		case <-ctx.Done():
			return
		}
	}
}

// Route function runs destination writer by mapping DestinationID from message to registered writer.
func (r *Relayer) route(msgs []*message.Message) {
	ctxWithSpan, span := otel.Tracer("relayer-core").Start(context.Background(), "relayer.core.Route")
	defer span.End()

	destChain, ok := r.registry[msgs[0].Destination]
	if !ok {
		log.Error().Msgf("no resolver for destID %v to send message registered", msgs[0].Destination)
		span.SetStatus(codes.Error, fmt.Sprintf("no resolver for destID %v to send message registered", msgs[0].Destination))
		return
	}

	log.Debug().Msgf("Routing %d messages to destination %d", len(msgs), destChain.DomainID())
	for _, m := range msgs {
		span.AddEvent("Routing message", traceapi.WithAttributes(attribute.String("msg.id", m.ID()), attribute.String("msg.type", string(m.Type))))
		log.Debug().Str("msg.id", m.ID()).Msgf("Routing message %+v", m.String())
		r.metrics.TrackDepositMessage(m)
		for _, mp := range r.messageProcessors {
			if err := mp(ctxWithSpan, m); err != nil {
				log.Error().Str("msg.id", m.ID()).Err(fmt.Errorf("error %w processing message %v", err, m.String()))
				return
			}
		}
	}

	err := destChain.Write(ctxWithSpan, msgs)
	if err != nil {
		for _, m := range msgs {
			log.Err(err).Str("msg.id", m.ID()).Msgf("Failed sending message %s to destination %v", m.String(), destChain.DomainID())
			r.metrics.TrackExecutionError(m)
		}
		return
	}

	for _, m := range msgs {
		r.metrics.TrackSuccessfulExecutionLatency(m)
	}
	span.SetStatus(codes.Ok, "messages routed")
}

func (r *Relayer) addRelayedChain(c RelayedChain) {
	if r.registry == nil {
		r.registry = make(map[uint8]RelayedChain)
	}
	domainID := c.DomainID()
	r.registry[domainID] = c
}
