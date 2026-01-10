package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/gen2brain/beeep"
	"github.com/olebedev/when"
	"github.com/olebedev/when/rules/common"
	"github.com/olebedev/when/rules/en"
)

const (
	markName  = "GOLANG_CLI_REMINDER"
	markValue = "1"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Printf("Usage: %s <hh:mm> <message> <time>\n", os.Args[0])
		os.Exit(1)
	}

	now := time.Now()

	w := when.New(nil)
	w.Add(en.All...)
	w.Add(common.All...)

	t, err := w.Parse(os.Args[1], now)

	if err != nil {
		fmt.Println("Error parsing time: %v\n", err)
		os.Exit(2)
	}

	if t == nil {
		fmt.Printf("Invalid time format: %s\n", os.Args[1])
		os.Exit(2)
	}
	if now.After(t.Time) {
		fmt.Printf("Time is in the past: %s\n", os.Args[1])
		os.Exit(3)
	}

	diff := t.Time.Sub(now)
	if os.Getenv(markName) == markValue {
		time.Sleep(diff)
		message := strings.Join(os.Args[2:], " ")
		// Use empty string for icon - beeep will use default system icon
		err = beeep.Alert("Reminder", message, "")
		if err != nil {
			fmt.Printf("Notification error: %v\n", err)
			// Fallback: print to console if notification fails
			fmt.Printf("REMINDER: %s\n", message)
			os.Exit(4)
		} else {
			fmt.Printf("Notification sent: %s\n", message)
		}
	} else {
		cmd := exec.Command(os.Args[0], os.Args[1:]...)
		cmd.Env = append(os.Environ(), fmt.Sprintf("%s=%s", markName, markValue))
		if err := cmd.Start(); err != nil {
			fmt.Println(err)
			os.Exit(5)
		}
		fmt.Println("Reminder set. Will notify after", diff.Round(time.Second))
		os.Exit(0)
	}
}
