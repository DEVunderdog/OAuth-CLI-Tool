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
	Use: "exit",
	Short: "Gracefully exit",
	Run: runExit,
}

func runExit(cmd *cobra.Command, args []string) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		for i := 0; i < 10; i ++ {
			select {
			case <-ctx.Done():
				return
			default:
				fmt.Printf("Working...%d\n", i)
				time.Sleep(1 * time.Second)
			}
		}
	}()

	<-ctx.Done()

	stop()

	fmt.Println("Exiting gracefully, press Ctrl + C again to force")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cleanup(shutdownCtx)

	fmt.Println("Bye!!! :)")
}

func cleanup(ctx context.Context) {
	for i := 0; i < 5; i++ {
		select {
		case <-ctx.Done():
			fmt.Println("Cleanup interrupted")
			return
		default:
			fmt.Printf("Cleaning up... %d\n", i)
			time.Sleep(1 * time.Second)
		}
	}
}
