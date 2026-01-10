# CLI Reminder in Golang

A simple command-line tool for setting reminders that send desktop notifications.

## Overview

This program lets you set a reminder by providing a time and a message. When the specified time arrives, it will show a desktop notification with your message.

## How it works

The program takes a time in `hh:mm` format (like 18:50 for 6:50 PM) and a message. It figures out how long to wait until that time, then starts a background process that sleeps until then. Once the time comes, it displays a notification on your desktop.

The background process part is handled by setting an environment variable. The first time you run it, it spawns a new process with that variable set, and that second process is the one that actually waits and shows the notification.

## Requirements

- Go 1.23.4 or later

## Installation

First install the dependencies:
```bash
go mod download
```

## Building

To build for Windows (from WSL or Linux):
```bash
GOOS=windows GOARCH=amd64 go build -o reminder.exe main.go
```

To build for Linux or macOS:
```bash
go build -o reminder main.go
```

## Usage

Run the program with a time and your message:
```bash
reminder.exe <hh:mm> <message>
```

Example:
```bash
reminder.exe 18:50 "Meeting with team"
reminder.exe 09:00 "Don't forget to call"
```

After running, you'll see a message like "Reminder set. Will notify after 2h30m" and then the program exits. The reminder runs in the background.

## Packages Used

- `github.com/gen2brain/beeep` - Handles showing desktop notifications on Windows, Linux, and macOS
- `github.com/olebedev/when` - Parses time strings into actual time values

## Limitations

- Time must be in 24-hour format (hh:mm)
- Can't set reminders for times that have already passed
- If you're using WSL on Windows, you need to build a Windows binary and run it from PowerShell, not from WSL
