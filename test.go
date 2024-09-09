package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	isTruped := make(chan struct{})
	go func() {
		defer close(isTruped)
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
		s := <-sig
		fmt.Printf("Signal received: %s \n", s.String())
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sleep", "60")
	cmd.Cancel = func() error {
		fmt.Println("hoge")
		return nil
	}
	cmd.WaitDelay = 5 * time.Second

	err := cmd.Start()
	if err != nil {
		fmt.Println(err)
		return
	}
	done := make(chan error, 1)
	go func() {
		if err := cmd.Wait(); err != nil {
			close(done)
			return
		}
		done <- nil
	}()
	go func() {
		select {
		case <-isTruped:
			fmt.Println("truped")
			cancel()
		}
	}()
	select {
	case _, ok := <-done:
		if ok {
			fmt.Println("done")
		} else {
			fmt.Println("closed")
		}
	}
}
