.PHONY: tailwind
tailwind:
	tailwindcss -o ./public/styles.css

.PHONY: templ
templ:
	templ generate

.PHONY: test
test:
	go test -race -v -timeout 30s ./...
