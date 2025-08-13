package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"meshtastic-exporter/pkg/factory"
	"meshtastic-exporter/pkg/hooks"
)

func TestNewMeshtasticHookSimple_Creation(t *testing.T) {
	// Этот тест проверяет, что NewMeshtasticHookSimple создается без ошибок
	// Реальная проверка TopicPrefix происходит в TestTopicFiltering_E2E

	f := factory.NewDefaultFactory()
	hook := hooks.NewMeshtasticHookSimple(f)

	// Проверяем, что хук создался
	assert.NotNil(t, hook, "Hook should be created")
	assert.Equal(t, "meshtastic", hook.ID(), "Hook should have correct ID")
}
