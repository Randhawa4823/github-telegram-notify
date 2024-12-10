# GitHub Webhook Listener

A simple GitHub webhook listener written in Go. It listens for incoming webhook requests from GitHub and handles various
types of GitHub events (e.g. push, pull request, issue) and sends them to a Telegram chat.

## Overview

The GitHub Webhook Listener project is a simple application written in Go that listens for incoming webhook requests
from GitHub. It handles various types of GitHub events such as push, pull request, issue, and more. The application is
designed to be a basic example of how to create a GitHub webhook listener using Go.

## Demo Bot
[FallenAlert](https://t.me/FallenAlertBot) || https://git-hook.vercel.app/

## Features

* Listens for incoming webhook requests from GitHub
* Handles various types of GitHub events (e.g. push, pull request, issue)
* Serves a simple HTML page at the root URL

## Requirements

* Go version 1.23.3 or higher
* BotToken - Get it from [Telegram](https://t.me/BotFather)
* GitHub webhook setup (see below for details)
* Vercel if you want to deploy it to production
* Ngrok if you want to deploy it to localhost

## Installation

1. Install Go (if not already installed) : [https://go.dev/dl/](https://go.dev/dl/)
2. Clone this repository to your local machine and navigate to the project directory
3. Run `go run main.go` to start the application
4. Configure your GitHub webhook to point to `http://localhost:3000/github` use Ngrok for testing/localhost

> As you know You can't use localhost for webhooks. you can
> use [Ngrok](https://dashboard.ngrok.com/get-started/setup/linux) for that.

## GitHub Webhook Setup

1. Go to your GitHub repository settings
2. Click on "Webhooks"
3. Click on "Add webhook"
4. Enter the URL `http://localhost:3000/github?chat_id=-1000000000`
5. Set Content type to "application/json"
6. Choose the events you want to listen for (e.g. push, pull request, issue)
7. Click "Add webhook"

## Deployment

### Vercel Deployment

1. Fork this repository to your GitHub account
2. Visit Vercel.com and Create a new Vercel project
3. Deploy the forked repository to Vercel
4. Done !

## API Documentation

* `/github`: Handles incoming webhook requests from GitHub (e.g. `/github?chat_id=123456789`)
* `/`: Serves a simple HTML page

## Contributing

Pull requests are welcome! If you'd like to contribute to this project, please fork the repository and submit a pull
request.

## License

This project is licensed under the [MIT License](LICENSE).
