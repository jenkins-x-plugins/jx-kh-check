# jx-bot-token

Will validate the Jenkins X pipeline bot token returns a success response code after hitting a git provider endpoint.

Environment variables:

	OAUTH_TOKEN
	GIT_PROVIDER

Currently only `https://github.com` provider is supported but more will be added 