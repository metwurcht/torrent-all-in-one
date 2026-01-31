package cli

import "fmt"

// ConsoleReporter affiche la progression dans la console
type ConsoleReporter struct{}

func (c *ConsoleReporter) OnStart(message string) {
	fmt.Println(message)
}

func (c *ConsoleReporter) OnProgress(message string) {
	fmt.Println(message)
}

func (c *ConsoleReporter) OnComplete(message string) {
	fmt.Println(message)
}

func (c *ConsoleReporter) OnError(message string) {
	fmt.Println(message)
}
