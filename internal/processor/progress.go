package processor

// ProgressReporter permet au processor de notifier sa progression
type ProgressReporter interface {
	// OnStart est appelé au début du traitement
	OnStart(message string)

	// OnProgress est appelé pendant le traitement
	OnProgress(message string)

	// OnComplete est appelé quand une étape est terminée
	OnComplete(message string)

	// OnError est appelé en cas d'erreur
	OnError(message string)
}

// SilentReporter est un reporter qui ne fait rien (pour les tests ou usage programmatique)
type SilentReporter struct{}

func (s *SilentReporter) OnStart(message string)    {}
func (s *SilentReporter) OnProgress(message string) {}
func (s *SilentReporter) OnComplete(message string) {}
func (s *SilentReporter) OnError(message string)    {}
