.PHONY: dev

.PHONY: dev
dev:
	air

.PHONY: gen
gen:
	templ generate

.PHONY: install
install:
	go install github.com/cosmtrek/air@latest
	go install github.com/a-h/templ/cmd/templ@latest

.PHONY: css
css:
	bun run tailwindcss -i ./assets/css/style.css -o ./ui/style.css --watch

.PHONY: changelog
changelog:
	git-chglog -o CHANGELOG.md --next-tag ${TAG}
