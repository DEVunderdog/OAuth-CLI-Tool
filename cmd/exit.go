package cmd

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

var ExitCmd = &cobra.Command{
	Use:   "exit",
	Short: "Gracefully exit",
	Run:   runExit,
}

func runExit(cmd *cobra.Command, args []string) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		for i := 0; i < 10; i++ {
			select {
			case <-ctx.Done():
				return
			default:
				time.Sleep(1 * time.Second)
			}
		}
	}()

	fmt.Println("Need to exit, press Ctrl + C")
	<-ctx.Done()

	stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cleanup(shutdownCtx)

	fmt.Println("\nBye!!! :)")
}

func cleanup(ctx context.Context) {
	for i := 0; i < 5; i++ {
		select {
		case <-ctx.Done():
			fmt.Println("Cleanup interrupted")
			return
		default:
			fmt.Printf("\nCleaning up... %d", i)
			time.Sleep(1 * time.Second)
		}
	}
}
