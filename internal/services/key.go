package services

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log/slog"

	"github.com/hesoyamTM/apphelper-schedule/internal/lib/encoding"
)

type KeyManager struct {
	log   *slog.Logger
	keyCh chan *ecdsa.PublicKey
}

func New(log *slog.Logger) *KeyManager {
	km := &KeyManager{log: log}

	km.keyCh = make(chan *ecdsa.PublicKey, 1)

	return km
}

func (k *KeyManager) SetPublicKey(ctx context.Context, key string) error {
	const op = "key.SetPublicKey"
	log := k.log.With(slog.String("op", op))

	pubKey, err := encoding.DecodeKey(key)
	if err != nil {
		log.Error("failed to decode key", slog.String("Error", err.Error()))
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Debug("received new public key")

	k.keyCh <- pubKey

	return nil
}

func (k *KeyManager) GetKeyChannel() <-chan *ecdsa.PublicKey {
	return k.keyCh
}
