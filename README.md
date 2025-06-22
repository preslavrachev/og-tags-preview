# Open Graph tags url preview service
A simple API that extracts and returns [Open Graph](https://ogp.me/) tags from a given URL in JSON format. Includes caching using Redis, circuit breakers, error handling, metrics monitoring with Prometheus, and a reverse proxy using Caddy.

It won't work for sites that prevent bots like Reddit, X(Twitter), etc.

###### Motivation
This project is part of a series of backend projects from [Stream](https://stream-wiki.notion.site/Stream-Go-10-Week-Backend-Eng-Onboarding-625363c8c3684753b7f2b7d829bcd67a#ec1a331b6b79470a8861449c13c895d0). I thought these were really good projects as they cover many relevant aspects of backend engineering. I built them as thoughtfully as I could and sought feedback from more experienced engineers. If you have any suggestions or improvements, I’d love to hear them.

## System Overview
![Preview](https://i.ibb.co/mFdRL2Vv/Blank-document.jpg)

## Quick Start
1. Clone the repo
2. Create a `.env` file in root directory with configurations
    ```
    PORT=4000
    ENV=local
    SERVER_IDLETIMEOUT=1
    SERVER_READTIMEOUT=5
    SERVER_WRITETIMEOUT=5
    REDIS_ADDR=localhost:6379
    REDIS_PASSWORD=
    REDIS_DB=0
    ```
3. Start redis listen at configured `addr`
4. `go mod tidy`
5. `make run`

## Usage
Send a `POST` request to the api's `/og` path with the `url` in `json` format
```
curl -X POST http://localhost:4000/og \
  -H "Content-Type: application/json" \
  -d '{"url": "https://ogp.me/"}'
```
If everything run correctly you should get
```
trung ~/Desktop$ curl -X POST http://localhost:4000/og \
  -H "Content-Type: application/json" \
  -d '{"url": "https://ogp.me/"}'
{
	"result": {
		"url": "https://ogp.me/",
		"og_tags": [
			"og:title Open Graph protocol",
			"og:type website",
			"og:url https://ogp.me/",
			"og:image https://ogp.me/logo.png",
			"og:image:type image/png",
			"og:image:width 300",
			"og:image:height 300",
			"og:image:alt The Open Graph logo",
			"og:description The Open Graph protocol enables any web page to become a rich object in a social graph."
		]
	}
}
```

## Contributing
If you have any suggestions, feedbacks, bug reports, feel free the share. If you want to contribute just create an issue and make a PR to `main` :)
