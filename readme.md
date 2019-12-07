# Search-O-Rama POC

This is an implementation for a webcrawler/search index. 

Since it's a POC code, the emphasis was on connecting all the dots, and creating a compelling-enough/correct-enough story to get feedback from key stackholders. 

Some questions that need feedback, are:
- how to deal with errors? (I implemented a partial-success error pattern, like GQL does, but this may not be the desired behavior)
- how to handle re-indexing a page?
- how to treat bad text?
- what non-alphanumeric terms need to be kept?
- what are other things that need to be done with the data post-crawl/indexing? (This will affect how the datamodels evolve)
- etc

Requires: go 1.12

`go get ./...`

`go test`

`go run .`

Browse to: `localhost:7250` and play!

A basic UI is provided to kick the tires. 

There's plenty of room for improvement, but that's the case with any code that's written!

-----

The webserver runs on: `localhost:7250/api`

The REST interface accepts:

`POST /api/crawl`:

- `Request Body: {"URLString": "http://wikipedia.org"}`
- `Response Body: {"DurationSeconds":2.982126938,"PagesCrawled":25,"WordsIndexed":4705,"CrawlErrors":null (or array of string)}`

`POST /api/search`:

- `Request Body: {"Term": "cat"}`
- `Response Body: {"DurationSeconds":0.000120686,"Results":[{"URL":"http://example.com/","Term":"cat","Title":"we love cats","Count":240}`

`DELETE /api/reset`
- no request or response body


Tests
-----
The test code is integration level to prove out that the business requirements were met, without creating alot of test-code that may change once feedback is acquired. 

Once feedback is obtained, more finegrained tests will be implemented to ensure that everything stays covered as the codebase goes into production.

----

Architecture
----
POC code has a tendency to turn in production code in short order, so I tried to balance keeping decent architectural boundaries, while keeping with a quick-and-dirty theme.

Data Bits:

The main data storage/repository is handled in these files:
- `page.go term.go`

The external facing data-structures are in:
- `contracts.go`

(I didn't spend a ton of time creating pretty JSON tags, but that's polish that can be added later)

The meat and potatoes are in:
- `crawler.go`

The webserver/static file server and all the handlers are in:
- `main.go`

I've gated the concurrency for a given crawl/index operation to 15 pages at a time, and the link-depth to crawl is gated at 3.


-----
